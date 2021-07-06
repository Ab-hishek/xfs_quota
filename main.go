package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"time"
)

var (
	dir                = "/xfs-dir/xfs_disk/"
	podVolumeMountPath = "/example"
	limit              = "5m"
	dirNotCreated      = true
	sleepTime          = 180
)

const (
	xfsQuota = "xfs_quota"
	mkdir    = "mkdir"
	mkfs     = "mkfs"
	dd       = "dd"
	mount    = "mount"
)

func main() {
	// Create sparse file using seek, 32mb max size
	args := []string{"if=/dev/zero", "of=" + podVolumeMountPath + "/xfs7.32M", "bs=1", "count=0", "seek=32M"}
	err := Run(dd, args)
	if err != nil {
		log.Fatalf("Error creating sparse file using seek: %+v", err)
	}
	log.Println("Successfully created sparse file.")

	// format in xfs format
	args = []string{"-t", "xfs", "-f", "-q", podVolumeMountPath + "/xfs7.32M"}
	err = Run(mkfs, args)
	if err != nil {
		log.Fatalf("Error formatting file in xfs format: %+v", err)
	}
	log.Println("Successfully formatted the sparse file.")

	// create a directory where mount will occur
	args = []string{"-p", podVolumeMountPath + dir}
	err = Run(mkdir, args)
	if err != nil {
		log.Fatalf("Error creating a mount directory: %+v", err)
	}
	log.Println("Successfully created hostpath directory.")

	// mount as loopback device with project quota enabled
	args = []string{"-o", "loop,rw,pquota", podVolumeMountPath + "/xfs7.32M", podVolumeMountPath + dir}
	err = Run(mount, args)
	if err != nil {
		log.Fatalf("Error mounting loopback device with project quota: %+v", err)
	}
	log.Println("Successfully mounted loopback device with quota.")

	pvcName := ""
	log.Println("Looking for the hostpath directory created by OpenEBS...")
	for dirNotCreated {
		// Get the list of directories inside the hostpath created dir
		files, err := getSubdirectories(podVolumeMountPath + dir)
		if err != nil {
			log.Fatalf("Error getting directory list: %+v", err)
		}

		if len(files) > 0 {
			// Matching pattern
			// Using MatchString() function
			match, err := regexp.MatchString("pvc.*", files[0].Name())
			if err != nil {
				log.Fatalf("Directory name doesn't satisfy the matching criteria(i.e pvc.*): %+v", err)
			}
			if match {
				pvcName = files[0].Name()
				log.Println("Found hostpath volume created by OpenEBS.")
				dirNotCreated = false
			}
		} else {
			time.Sleep(time.Duration(sleepTime) * time.Second)
		}
	}

	var id string
	id = "500"
	// initialise project
	args = []string{"-x", "-c", fmt.Sprintf("%s%s%s%s", "'project -s -p ", podVolumeMountPath+dir+pvcName, " ", id+"'")}
	err = Run(xfsQuota, args)
	if err != nil {
		log.Fatalf("Error initializing project: %+v", err)
	}
	log.Println("Successfully created new sfs project.")

	// set a 5m quota on project, id =100
	args = []string{"-x", "-c", fmt.Sprintf("%s%s%s", "'limit -p ", "bsoft="+limit+" bhard="+limit, id+"'"), podVolumeMountPath + dir}
	err = Run(xfsQuota, args)
	if err != nil {
		log.Fatalf("Error seeting project quota: %+v", err)
	}
	log.Println("Successfully set quota onto the volume.")

}

func Run(command string, args []string) error {
	cmd := exec.Command(command, args...)
	log.Printf("Args: %+v", cmd.Args)
	log.Printf(cmd.String())
	_, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	return nil
}

// fileExists checks if a dir exists or not before we
// try using it to prevent further errors.
/*func fileExists(filename string) (bool, error) {
	matches, err := filepath.Glob(filename + ".*")
	if err != nil {
		return false, err
	}
	return len(matches) > 0, nil
}*/

func getSubdirectories(pathName string) ([]os.FileInfo, error) {
	files, err := ioutil.ReadDir(pathName)
	if err != nil {
		return nil, err
	}

	return files, nil
}
