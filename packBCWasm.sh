#!/bin/bash
TAG=$1
OUTPUT_DIR=$2

showUsage() {
  echo 
  echo "Usage:"
  echo "./packWASM.sh <release_tag> [output_dir]"
  echo "example: ./packWASM.sh 0.1.4.4.5 .."
  echo "[output_dir] can be ommitted, default output dir is ../"
}

[ "$TAG" == "" ] && echo && echo "no release tag set!" && showUsage && exit 0
[ "$OUTPUT_DIR" == "" ] && OUTPUT_DIR=..
#[ "$GOPATH" == "" ] && echo "GOPATH not set!" && exit
echo "HINT: default output dir is ../, you can set the dir like \"./packWASM.sh release_tag your_custom_output_dir.\""

#cd $GOPATH/src/github.com/PlatONEnetwork/PlatONE-Go
[ -d ${OUTPUT_DIR}/SysContracts ] && echo "SysContracts exsits in ${OUTPUT_DIR}, please remove it first." && exit
cp -r cmd/SysContracts ${OUTPUT_DIR}
cd $OUTPUT_DIR
[ -f BCWasm-linux_release*.tar.gz ] && rm  BCWasm-linux_release*.tar.gz
[ -d BCWasm ] && rm -rf BCWasm
mv SysContracts BCWasm
cp PlatONE-Go/release/linux/bin/ctool BCWasm/external/bin/
rm -rf BCWasm/systemContract
rm -rf BCWasm/build
tar -zcf BCWasm-linux_release.${TAG}.tar.gz BCWasm
rm -rf BCWasm
