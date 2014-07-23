#! /bin/bash
base=github.com/godfried/impendulo/
cd $IMPENDULO
dirs=$(find . -mindepth 1 -maxdepth 2 -type d  \( ! -iname ".*" \) | sed 's|^\./||g')
for d in $dirs
do  
    if [[ $d != .* ]] && [[ $d != static* ]] && [[ $d != java* ]] && [[ $d != scripts* ]]
	then
	go test -cover "$base""$d"
	fi
done
