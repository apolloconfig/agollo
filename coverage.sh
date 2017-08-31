set -e

workdir=.cover
profile="$workdir/cover.out"
mode=count

generate_cover_data() {
    rm -rf "$workdir"
    mkdir "$workdir"

    for pkg in "$@"; do
        f="$workdir/$(echo $pkg | tr / -).cover"
        go test -covermode="$mode" -coverprofile="$f" "$pkg"
    done

    echo "mode: $mode" >"$profile"
    grep -h -v "^mode:" "$workdir"/*.cover >>"$profile"
}

show_html_report() {
    go tool cover -html="$profile" -o="$workdir"/coverage.html
}

show_csv_report() {
    go tool cover -func="$profile" -o="$workdir"/coverage.csv
}

push_to_coveralls() {
    echo "Pushing coverage statistics to coveralls.io"
    # ignore failure to push - it happens
    $HOME/gopath/bin/goveralls -coverprofile="$profile" \
                               -service=travis-ci  || true
}

generate_cover_data $(go list ./...)
show_csv_report

case "$1" in
"")
    ;;
--html)
    show_html_report ;;
--coveralls)
    push_to_coveralls ;;
*)
    echo >&2 "error: invalid option: $1"; exit 1 ;;
esac