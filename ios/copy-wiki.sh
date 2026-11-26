#!/bin/zsh
# Copies the repo's wiki into the app bundle at build time: content only.
# The learner's progress, backups, crash log and the private folder
# (copyrighted episodes, the frequency list) never go into the app.
set -eu
SRC="${SRCROOT}/../wiki/"
DEST="${TARGET_BUILD_DIR}/${UNLOCALIZED_RESOURCES_FOLDER_PATH}/wiki/"
mkdir -p "$DEST"
/usr/bin/rsync -a --delete --delete-excluded \
  --exclude '.DS_Store' --exclude 'private/' --exclude 'progress.db*' --exclude 'backups/' \
  --exclude '*.imported' --exclude '*.adopted' --exclude 'saves/' --exclude 'sessions/' \
  --exclude 'crash.log' --exclude 'images/' \
  --exclude '/journey.yaml' --exclude '/study.yaml' --exclude '/word-schedule.yaml' \
  --exclude '/settings.yaml' --exclude '/forge.yaml' \
  "$SRC" "$DEST"
