# Build for s2i python images

# Release process

e.g. from 0.3-SNAPSHOT to release 0.3 and create 0.4-SNAPSHOT

 * set IMAGE_VERSION to new stable version X in Makefile (e.g. 0.3)
  * ```cd build_scripts``` and run ```./build_all.sh``` and then ```./push_all.sh```
 * Update IMAGE_VERSION to (X+1)-SNAPSHOT (e.g. 0.4-SNAPSHOT)
  * ```./build_all.sh``` and then ```./push_all.sh```
 * Update main readme to show new versions of stable and snapshot
 * Update versions in docs, Makefiles and notebooks of stable version
    ``` ./update_python_version.sh X X+1```, e.g ```./update_python_version.sh 0.2 0.3```
    (check that any references to X-SNAPSHOT are updated to X - may need manual intervention)
 * Update the `doc/source/reference/images.md` manually with the latest image versions
