# Move ephmatches to ephmatch_likes
INSERT INTO ephmatch_likes
(id, created_at, updated_at, user_id, liked_id)
SELECT id, created_at, updated_at, user_id, other_id
FROM ephmatches;

# Move matching ephmatches to ephmatch_matches (does all matches bc ephmatch og double counts
INSERT INTO ephmatch_matches
(created_at, updated_at, user_a_id, user_b_id)
SELECT a.created_at, a.updated_at, a.user_id, a.other_id
FROM ephmatches AS a
         INNER JOIN ephmatches AS b ON b.other_id = a.user_id
WHERE
        a.other_id=b.user_id AND
        a.user_id < a.other_id;

# Load user home data into locations
UPDATE
    ephmatch_profiles,
    users
SET
    ephmatch_profiles.location_visible = users.home_visible,
    ephmatch_profiles.location_town = users.home_town,
    ephmatch_profiles.location_state = users.home_state,
    ephmatch_profiles.location_country = users.home_country
WHERE ephmatch_profiles.user_id = users.id;