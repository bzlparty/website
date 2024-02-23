curl \
  -H "Accept: application/vnd.github+json" \
  -H "X-GitHub-Api-Version: 2022-11-28" \
  -L \
  https://api.github.com/repos/bzlparty/$1/tarball/$2 |\
  tar xvzf - --strip-components=1 --wildcards **/*.md
