start ~ sudo -E env PATH="$PATH" sh -c "docker-compose up"

logs ~ sudo strace -p 2299402 &> loga.txt

pstree ~ pstree -p 2304832