# Generate the Go protobuf using the Docker build (writes pb file to repo root).
protogen:
	docker build --output type=local,dest=. --file Dockerfile.generate .
	@ls -l testproto/testproto.pb.go
