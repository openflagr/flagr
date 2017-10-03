main() {
    IFS=':' read -r -a paths <<< "$GOPATH"
    for path in "${paths[@]}"; do
        flagr_path="$path/src/github.com/checkr/flagr"
        if [ -d "$flagr_path" ]; then
            if [ "$flagr_path" -ef "$PWD" ]; then
               exit 0
            fi
        fi
    done

    echo "ERROR"
    echo "Project not found in ${GOPATH}."
    exit 1
}

main
