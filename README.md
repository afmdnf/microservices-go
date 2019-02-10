# Microservices in Go

In single node applications, we deal with complexity through libraries and functions. However, these are all linked together at compile time and can't change while the application is running. To help manage this complexity and to allow different teams to develop their modules independently, many organizations organize their applications around microservices.

Microservices are small programs with a well-defined interface that rarely changes. If we need more scale, we simply run more copies of the program on different nodes. If there is a new feature or bug fix, we can roll it out slowly by replacing only a few of the copies at a time. Finally, this type of deployment can be tolerant of faults since each service is only responsible for a limited part of the whole application, and there are usually many copies of each service to chose from if something goes wrong.

Microservices also provide a form of concurrency. In this context, concurrency refers to doing many different tasks at the same time (so here, the microservices would be performing concurrently). Owing to Golang's excellent support for concurrency, this project is implemented in Go and goroutines/channels are utilized for providing concurrency.

<p align="center">
  <img src="https://github.com/afmdnf/microservices-go/blob/master/overview.png" width="400">
  <br/>
</p>

**In this project, a memoization service for a machine learning classifier is created. Clients will send the service a request to classify an MNIST image. If the image is unseen, the machine-learning microservice is asked to classify it, but ML models can be expensive and slow to run. To speed things up, the service will save the classification of that image in a caching service. The next time the same image is requested, the answer can be fetched from the cache instead of the ML model, improving average performance significantly.**

## Classifier Service

<p align="center">
  <img src="https://github.com/afmdnf/microservices-go/blob/master/mnist.png" width="50">
  <br/>
</p>

The classifier service is designed to translate hand-written digits from the [classic MNIST dataset](https://en.wikipedia.org/wiki/MNIST_database). The input to it is a `[]byte` which represents a 28x28 pixel image. The classifier then runs an ensemble of support vector machine (SVM) classifiers, one for each digit, and picks the most likely digit to return (as an int). Like every service in this project, the classifier accepts a request ID with every request and returns that ID with the corresponding response. This allows for potentially out-of-order messages (which can occur due to network issues or from optimizations).

## Cache Service
The caching service is essentially a remote hash table. Caching services reduce the load on other services by saving frequent or recent results. The cache in this project is a very simple key-value store. To simplify things, it is assumed that it will never fill up (no eviction/replacement). You send it requests using the `CacheReq` struct which contains a read/write flag, a 64-bit key, an int value, and (like everything) a 64-bit requestID. If the request is a write, then the cache will not respond. If the request is a read, then the cache will respond with a `CacheResp` struct that contains the requestID, an `exists` flag (true if the item was found) and the corresponding value (if any).
Note that the key is a 64-bit integer, but the images used are byte slices. To deal with this, the hash of the image is used as a key.

## Memoization Layer
The memoization service uses the cache service to speed up requests to classify images that have already been seen before, in order to respond to them significantly faster. Existing users of the classifier should be able to place the memoization service in and immediately see benefits without changing any of their existing code.
