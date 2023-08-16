container: pircontainer oprfcontainer clientcontainer psikrssw19container cfcontainer


pircontainer:
	podman build -f Containerfile.pir -t pir:1.0.0 ./..

oprfcontainer:
	podman build -f Containerfile.oprf -t oprf:1.0.0 ./..

cfcontainer:
	podman build -f Containerfile.cf -t cf:1.0.0 ./..

clientcontainer:
	podman build -f Containerfile.client -t client:1.0.0 ./..

psikrssw19container:
	podman build -f Containerfile.psikrssw19 -t psi-krssw19:1.0.0 ./..
