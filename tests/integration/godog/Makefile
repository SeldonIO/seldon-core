test:
	go test --godog.tags='@ModelDeployment,@CustomModelSpec' --godog.concurrency=3 -race

debug:
	go test --godog.tags='@0' --godog.concurrency=1 -race

test-all:
	cd cicd && go test -race