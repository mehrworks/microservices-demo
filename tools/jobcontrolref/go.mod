module github.com/GoogleCloudPlatform/microservices-demo/tools/jobcontrolref

go 1.25

toolchain go1.25.6

require github.com/GoogleCloudPlatform/microservices-demo/src/frontend v0.0.0

replace github.com/GoogleCloudPlatform/microservices-demo/src/frontend => ../../src/frontend
