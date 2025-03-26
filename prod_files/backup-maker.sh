#!/bin/bash -e
##### WSO-Backend WSO 2.0 #####
# Automatically make backups.
# This script takes one argument which is a file that contains a list of
# relative paths for the program to archive. Leave it out for a full snapshot.
# The file backup-partial.txt contains what's backed up normally.

# DO NOT PUT COMMENTS IN BACKUP-PARTIAL! ENSURE THAT ALL FILES IN THERE
# EXIST, OR TAR WILL FAIL AND THERE'LL BE NO BACKUP!
# tar can gracefully recover sometimes. but don't rely on it! tar is a cruel mistress.

#    (__)    )
#    (..)   /|\
#   (o_o)  / | \
#   ___) \/,-|,-\
# //,-/_\ )  '  '
#    (//,-'\
#    (  ( . \_
#     `._\(___`.
#      '---' _)/
#           `-'
# Hic sunt dracones. Understanding or editing this is not advised.

# paths
SOURCE="/backup-sync"
BACKUP="/backup"
# matches the newest available cPanel-generated backup
TARGET=$(find /backup-sync -maxdepth 1 -type d -name "????-??-??" -exec stat --format="%Y %n" {} + | sort -n | tail -n 1 | cut -d ' ' -f 2-)
# use Windows-compatible filenames
TIMESTAMP=$(date -Iseconds | tr ':' '_')
# we use zstd compression since it strikes a nice balance between speed and
# archive size. these files are really big and the server is quite slow
BACKUP_OUTPUT_FILE="$BACKUP/wso-backup-$TIMESTAMP.tar.zst"

usage() {
  echo "Usage: $0 [-h] [-f file]";
  exit 0;
}
# arguments:
while getopts "vhf:" OPT; do
  case $OPT in
    h)
      usage
      ;;
    f)
      # file list argument
      if [ -z "$OPTARG" ]; then
        echo "(backup) error: '-f' requires a file argument."
        usage
        exit 1
      fi
      FILE_LIST="$OPTARG"
      if [ ! -f "$FILE_LIST" ]; then
        echo "(backup) error: file list '$FILE_LIST' not found."
        echo "(backup) are you sure it has the right permissions?"
        exit 1
      fi
      # this is so mind-numbingly stupid but Bash doesn't give other methods for concatenation
      TMPAFFIX=".tmp"
      sed "s|^|$TARGET/|" "$FILE_LIST" > "$FILE_LIST$TMPAFFIX"
      # note that this is technically a bash array, and not a string
      BACKUP_FILES=(-T "$FILE_LIST$TMPAFFIX")
      ;;
    *)
      usage
      ;;
  esac
done

echo "(backup) backup started at time: $TIMESTAMP"
if [ -z "$FILE_LIST" ]; then
  echo "(backup) no file list provided, archiving everything in $SOURCE"
  # tar syntax to nab everything
  BACKUP_FILES=(.)
fi

# estimate size of our compressed backup
# we can't always assume that compression will be more efficient or significant,
# so as a result we'll use the existing directory size as an estimate.
if [ -n "$FILE_LIST" ]; then
  # convert the file list into null-separated format for du, then summarize in bytes
  GUESS_SIZE=$(du -c --files0-from=<(tr '\n' '\0' < "${BACKUP_FILES[-1]}") --block-size=1 2>/dev/null | tail -n 1 | awk '{print $1}')
else
    # convert the file list into null-separated format for du, then summarize in bytes
  GUESS_SIZE=$(du -c --block-size=1 "$BACKUP" 2>/dev/null | tail -n 1 | awk '{print $1}')
fi
FREE_SIZE=$(df --output=avail --block-size=1 "$BACKUP" | tail -n 1)

# delete the oldest backup
rm_old_backup() {
  echo "(backup) not enough space to make next backup, deleting oldest..."
  while [ "$FREE_SIZE" -lt "$GUESS_SIZE" ]; do
    OLDEST_FILE=$(ls -tr "$BACKUP"/*.tar.zst 2>/dev/null | head -n 1)
    if [ -z "$OLDEST_FILE" ]; then
      echo "(backup) no more oldest backups."
      echo "(backup) something else is taking up storage."
      exit 1
    fi
    echo "(backup) deleting oldest file..."
    rm -rf "$OLDEST_FILE"
    # need to redefine this since storage has changed
    FREE_SIZE=$(df --output=avail --block-size=1 "$BACKUP_DIR" | tail -n 1)
  done
}

# see if we even need to delete older ones
if  [ "$FREE_SIZE" -lt "$GUESS_SIZE" ]; then
  rm_old_backup
fi

# create the compressed backup
echo "(backup) creating backup: $BACKUP_OUTPUT_FILE"

# we maintain all xattrs and strip all Windows-unsafe file names
# of course, this doesn't help for any files which are inside of the archive
# why is $SOURCE not in quotes? this is because something is wrong with tar's parsing
# I give up on trying to fix it. this is a gross hack.
tar -cf "$BACKUP_OUTPUT_FILE"  --ignore-failed-read --zstd --checkpoint-action=dot --xattrs-include='*.*' --numeric-owner --transform='s#[/:*?"<>|]#_#g' -C $SOURCE "${BACKUP_FILES[@]}" > /dev/null

# time has passed
TIMESTAMP=$(date -Iseconds | tr ':' '_')
# delete that temp file we made to overcome a Bash limitation
rm "$FILE_LIST$TMPAFFIX"
echo "(backup) backup complete at time: $TIMESTAMP"  
