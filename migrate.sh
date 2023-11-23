set -ue

db=euperturbot.db
version=$(echo "PRAGMA user_version;" | sqlite3 $db)
echo "current db version: $version"

next () {
    file=db/migrations/$[$version + 1].sql

    if test -f $file
    then
        version=$[$version + 1]
        echo "RUN $file"
        cat $file | sqlite3 $db
        echo "PRAGMA user_version = $version;" | sqlite3 $db
        next
    else
        echo "on latest version: $version"
    fi
}

next