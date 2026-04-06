#!/bin/sh

DB_PATH="${SUI_DB_FOLDER:-/app/db}/pupmmesub.db"
if [ ! -f "$DB_PATH" ]; then
	./sub migrate
fi

exec ./sub "$@"
