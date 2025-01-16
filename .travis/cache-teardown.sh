if [ "${PLUGINS_VISIBILITY}" = "private" ]; then
  rm -rf ${HOME}/.cache/go-build
  mkdir -p ${HOME}/.cache/go-build
fi
