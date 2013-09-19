#! /bin/bash
base=github.com/godfried/impendulo/
cd $GOPATH/src/$base
dirs=$(find . -mindepth 1 -maxdepth 2 -type d  \( ! -iname ".*" \) | sed 's|^\./||g')
for d in $dirs
do  
    if [[ $d != .* ]] && [[ $d != static* ]] && [[ $d != java* ]] && [[ $d != scripts* ]]
	then
	go test -i "$base""$d"
	fi
done
for d in $dirs
do  
    if [[ $d != .* ]] && [[ $d != static* ]] && [[ $d != java* ]] && [[ $d != scripts* ]]
	then
	gocov test "$base""$d" | gocov report
	fi
done
