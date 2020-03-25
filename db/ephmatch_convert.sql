INSERT INTO ephmatch_likes
(id, created_at, updated_at, user_id, liked_id)
SELECT id, created_at, updated_at, user_id, other_id
FROM ephmatches;

INSERT INTO ephmatch_matches
(created_at, updated_at, user_a_id, user_b_id)
SELECT a.created_at, a.updated_at, a.user_id, a.other_id
FROM ephmatches AS a
         INNER JOIN ephmatches AS b ON b.other_id = a.user_id
WHERE
        a.other_id=b.user_id AND
        a.user_id < a.other_id;