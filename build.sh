build(){
    printf $1/$2" Building...%5s"
    go env -w GOARCH=$2
    go env -w GOOS=$1

    name=ntn
    if [ $1 == "windows" ]; then 
        name='ntn.exe'
    fi
	
	mkdir dist/$1
    go build -ldflags "-w -s" -o dist/$1/$name client/main.go client/ntn.go
    if [ $? == 0 ]; then
        echo -e "\t[ OK ]"
		cp client/config.yaml dist/$1
    else
        echo -e "\t[ ERROR ]"
    fi
}

rm -rf dist
mkdir dist
build 'windows' '386'
build 'linux' '386'
build 'darwin' 'amd64'

go env -w GOOS=`go env GOHOSTOS`
go env -w GOARCH=`go env GOHOSTARCH`
