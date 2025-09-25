psql "postgres://postgres:postgres@localhost:5432/chirpy"

sudo -u postgres createuser -s northcape

sudo -u postgres psql -d chirpy
sudo -u postgres

goose postgres "postgres://postgres:postgres@localhost:5432/chirpy" up
goose postgres "postgres://postgres:postgres@localhost:5432/chirpy" down

psql chirpy