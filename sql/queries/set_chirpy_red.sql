-- name: SetChirpyRed :exec
UPDATE users
SET 
    updated_at = NOW(),
    is_chirpy_red = $1
WHERE id = $2;