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
	dir                = "/data/volumes/xfs_hostpath/"
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
	args := []string{"if=/dev/zero", "of=" + podVolumeMountPath + "/xfs1.32M", "bs=1", "count=0", "seek=32M"}
	err := Run(dd, args)
	if err != nil {
		log.Fatalf("Error creating sparse file using seek: %+v", err)
	}
	log.Println("Successfully created sparse file.")

	// format in xfs format
	args = []string{"-t", "xfs", "-q", podVolumeMountPath + "/xfs1.32M"}
	err = Run(mkfs, args)
	if err != nil {
		log.Fatalf("Error formatting file in xfs format: %+v", err)
	}
	log.Println("Successfully formatted the sparse file.")

	// create a directory where mount will occur
	args = []string{"-p", dir}

	err = Run(mkdir, args)
	if err != nil {
		log.Fatalf("Error creating a mount directory: %+v", err)
	}
	log.Println("Successfully created hostpath directory.")

	// mount as loopback device with project quota enabled
	args = []string{"-o", "loop,rw", podVolumeMountPath + "/xfs1.32M", "-o", "pquota", dir}
	err = Run(mount, args)
	if err != nil {
		log.Fatalf("Error mounting loopback device with project quota: %+v", err)
	}
	log.Println("Successfully mounted loopback device with quota.")

	var files []os.FileInfo
	log.Println("Looking for the hostpath directory created by OpenEBS:")
	for dirNotCreated {
		// Get the list of directories inside the hostpath created dir
		files, err := getSubdirectories(dir)
		if err != nil {
			log.Fatalf("Error getting directory list: %+v", err)
		}

		if len(files) > 0 {
			dirNotCreated = false
		} else {
			time.Sleep(time.Duration(sleepTime))
		}
	}

	// Matching pattern
	// Using MatchString() function
	_, err = regexp.MatchString(files[0].Name(), "pvc.*")
	if err != nil {
		log.Fatalf("Directory name doesn't satisfy the matching criteria: %+v", err)
	}
	log.Println("Found hostpath volume created by OpenEBS.")

	var id string
	id = "100"
	// initialise project
	args = []string{"-x", "-c", fmt.Sprintf("%s%s%s%s", "'project -s -p ", dir+files[0].Name(), " ", id+"'")}
	err = Run(xfsQuota, args)
	if err != nil {
		log.Fatalf("Error initializing project: %+v", err)
	}
	log.Println("Successfully created new sfs project.")

	// set a 5m quota on project, id =100
	args = []string{"-x", "-c", fmt.Sprintf("%s%s%s%s", "'limit -p ", "bsoft="+limit+" bhard="+limit, id+"'", dir)}
	err = Run(xfsQuota, args)
	if err != nil {
		log.Fatalf("Error seeting project quota: %+v", err)
	}
	log.Println("Successfully set quota onto the volume.")

}

func Run(command string, args []string) error {
	cmd := exec.Command(command, args...)
	err := cmd.Run()
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
