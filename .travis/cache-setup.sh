if [ "${PLUGINS_VISIBILITY}" = "private" ]; then
  source .travis/cache-teardown.sh

  hash_private=`echo -n "MONGODB=${MONGODB}=MONGOSH=${MONGOSH}=PLUGINS_VISIBILITY=private" | sha256sum | cut -f1 -d' '`
  hash_public=`echo -n "MONGODB=${MONGODB}=MONGOSH=${MONGOSH}=PLUGINS_VISIBILITY=public" | sha256sum | cut -f1 -d' '`

  type travis_run_setup_casher | sed -e '1,3d' -e "s/${hash_private}/${hash_public}/g" -e '/\\ cache\\ add\\ /d' -e '$d' > .travis/cache-setup-private
  
  source .travis/cache-setup-private

  echo "Private cache ${hash_private} replaced with public cache ${hash_public}."
fi
