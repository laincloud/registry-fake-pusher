package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/laincloud/registry-fake-pusher/rfp"
	"github.com/laincloud/registry-fake-pusher/rfp/utils/log"
)

// Need use registry v2 API
// Precondition：
//    - srcRepository:sourceTag already exists in sourceRegistry
//    - all the imagelayers to be pushed exist in sourceRegistry
//    - targetRepository:targetTag already exists in targetRegistry
//    - newTag is a tag not exist in targetRegistry
func main() {

	var srcRegistry, srcRepository, srcTag, targetRegistry, targetRepository, targetTag, newTag string
	var isDebug bool
	var srcJWT, targetJWT string
	var srcLayerCount int

	flag.StringVar(&srcRegistry, "srcReg", "registry.example.com", "The domain of source regsitry")
	flag.StringVar(&srcRepository, "srcRepo", "sourceRepo", "The repository which exists an image layer you want to copy to other repository")
	flag.StringVar(&srcTag, "srcTag", "sourceTag", "The tag which exists image layer you want to copy to other repository")
	flag.IntVar(&srcLayerCount, "srcLayerCount", 1, "The layer count of source tag from top to overlay to target tag")
	flag.StringVar(&targetRegistry, "targetReg", "registry.example.com", "The domain of target regsitry")
	flag.StringVar(&targetRepository, "targetRepo", "targetRepo", "The repository which exist a tag you want to copy a layer to")
	flag.StringVar(&targetTag, "targetTag", "targetTag", "The tag which you want to copy a layer to")
	flag.StringVar(&newTag, "newTag", "newTag", "The tag been generated after the operation")
	flag.BoolVar(&isDebug, "debug", false, "Debug mode switch")
	flag.StringVar(&srcJWT, "srcJWT", "", "Optional! The JWT used to access the source registry and repository")
	flag.StringVar(&targetJWT, "targetJWT", "", "Optional! The JWT used to access the target registry and repository")
	flag.Parse()

	if isDebug {
		log.EnableDebug()
	}

	pusher, err := rfp.NewRegistryFakePusher(srcRegistry, srcRepository, srcTag, targetRegistry, targetRepository, targetTag, newTag)
	if err != nil {
		fmt.Println("Error when initial push : ", err)
		os.Exit(1)
	}

	if err := pusher.FakePush(srcJWT, targetJWT, srcLayerCount); err != nil {
		fmt.Println("Registry Fake Push failed: ", err)
		os.Exit(2)
	}

	os.Exit(0)

}
