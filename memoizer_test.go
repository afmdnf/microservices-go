package proj

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"testing"
	"time"

	"github.com/petar/GoMNIST"
)

const bufSize = 100

/* Initializes the whole system using the provided functions for the various services.
Returns: a handle to the memoizer */
func startup(memFunc func(MnistHandle, MnistHandle, CacheHandle),
	classFunc func(MnistHandle),
	cacheFunc func(CacheHandle),
) MnistHandle {
	/* Create all the services and their handles */
	classHandle := MnistHandle{
		make(chan MnistReq, bufSize),
		make(chan MnistResp, bufSize),
	}
	go classFunc(classHandle)

	cacheHandle := CacheHandle{
		make(chan CacheReq, bufSize),
		make(chan CacheResp, bufSize),
	}
	go cacheFunc(cacheHandle)

	memHandle := MnistHandle{
		make(chan MnistReq, bufSize),
		make(chan MnistResp, bufSize),
	}
	go memFunc(memHandle, classHandle, cacheHandle)

	return memHandle
}

// Minimal correctness testing
func TestMemoizerMinimal(t *testing.T) {
	rawTrain, err := GoMNIST.ReadSet(trainDataPath, trainLblPath)
	if err != nil {
		panic(fmt.Sprintf("Failed to load training set from %s and %s: %v\n",
			trainDataPath, trainLblPath, err))
	}

	// Used to ensure that message IDs are always globally unique
	var reqID int64 = 0

	/* To see what these images look like, uncomment this */
	// Show(rawTrain.Images[0])

	// Initialize our system for basic testing
	memHandle := startup(Memoizer, MnistServer, Cache)

	CheckImage(rawTrain.Images[0], int(rawTrain.Labels[0]), memHandle, &reqID, t)

	// == Make sure it works the same way twice in a row == */
	CheckImage(rawTrain.Images[0], int(rawTrain.Labels[0]), memHandle, &reqID, t)

	//== Test some more values, make sure the model is deterministic.
	// These tests should always pass, no matter the model since they only check
	// that the response is reasonable
	firstResp := CheckImage(rawTrain.Images[1], -1, memHandle, &reqID, t)
	if resp := CheckImage(rawTrain.Images[1], -1, memHandle, &reqID, t); resp != firstResp {
		t.Errorf("Classification on second attempt doesnt match first, %d != %d", firstResp, resp)
	}

	// == Close the channel
	close(memHandle.ReqQ)
}

/* This is how much faster the hot cache run should at least be. */
const cacheSpeedup = 2.0

// Tests if the memoizer is really memoizing (i.e. using the cache)
func TestMemoizerSpeedup(t *testing.T) {
	rawTrain, err := GoMNIST.ReadSet(trainDataPath, trainLblPath)
	if err != nil {
		panic(fmt.Sprintf("Failed to load training set from %s and %s: %v\n",
			trainDataPath, trainLblPath, err))
	}

	memHandle := startup(Memoizer, MnistServer, Cache)

	//Globally unique request ID
	var reqID int64 = 0

	// == Test a bunch of images that we've never seen before (cache can't help us)
	start := time.Now()
	CheckImages(rawTrain.Images[2:], nil, memHandle, &reqID, t)
	coldRunTime := time.Since(start)

	// == Now test again with the same images, the cache should help here
	start = time.Now()
	CheckImages(rawTrain.Images[2:], nil, memHandle, &reqID, t)
	hotRunTime := time.Since(start)

	// == The second run should be faster
	if coldRunTime/hotRunTime < cacheSpeedup {
		t.Errorf("The cache didn't seem to help. Cold run: %v, Hot run: %v", coldRunTime, hotRunTime)
	}
	fmt.Printf("Cold run took %v, Hot run took %v\n", coldRunTime, hotRunTime)

	// == Close the channel
	close(memHandle.ReqQ)
}

/* Will print the raw image to the current directory as "out.png" */
func Show(data GoMNIST.RawImage) {
	const dim = 28
	/* Create an empty image */
	m := image.NewNRGBA(image.Rect(0, 0, dim, dim))

	/* Copy the bytes into it (in grayscale) from the raw image */
	for y := 0; y < dim; y++ {
		for x := 0; x < dim; x++ {
			v := data[y*dim+x]
			m.Set(x, y, color.RGBA{v, v, v, 255})
		}
	}

	/* Write the image to a file. */
	f, _ := os.OpenFile("out.png", os.O_WRONLY|os.O_CREATE, 0600)
	defer f.Close()
	png.Encode(f, m)
}
