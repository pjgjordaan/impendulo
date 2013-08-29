#! /bin/bash
base=github.com/godfried/impendulo/
dirs=$(find . -mindepth 1 -maxdepth 2 -type d  \( ! -iname ".*" \) | sed 's|^\./||g')
for d in $dirs
do  
    if [[ $d != .* ]] && [[ $d != static* ]]
	then
	go test -i "$base""$d"
	fi
done
for d in $dirs
do  
    if [[ $d != .* ]] && [[ $d != static* ]]
	then
	go test "$base""$d"
	fi
done