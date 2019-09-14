# This is the file to convert the production WSO-on-Rails DB to WSO-Go
# We are assuming the databases are named wso_rails and wso_go, respectively
# We are also assuming wso_go has been created and contains all schemas required.

/*
 wso_rails tables:
     api_keys # ignore
     ar_internal_metadata # ignore
     areas_of_study
     book_listings # unimplemented
     books # unimplemented
     bulletins
     courses
     departments
     discussions
     dorm_rooms
     dorms
     dormtrak_reviews
     ephcatches
     factrak_agreements
     factrak_surveys
     neighborhoods
     offices
     posts
     schedulecourses # unimplemented
     schema_migrations # ignore
     tags
     tags_users
     users
     wms_courses # unimplemented
     workshops # unimplemented

 wso_go tables:
     areas_of_study
     bulletin_rides
     bulletins
     courses
     departments
     discussions
     dorm_rooms
     dorms
     dormtrak_reviews
     ephcatches
     factrak_agreements
     factrak_surveys
     migrations
     neighborhoods
     offices
     posts
     tags
     tags_users
     users
 */

# areas_of_study
# Query
INSERT INTO wso_go.areas_of_study
    (id, name, abbrev, department_id, created_at, updated_at)
    SELECT id, name, abbrev, department_id, NOW(), NOW()
FROM wso_rails.areas_of_study;


# bulletins
# Query
INSERT INTO wso_go.bulletins
    (id, type, title, body, start_date, offer, user_id, created_at, updated_at)
    SELECT id, type, title, body, start_date, offer, user_id, created_at, updated_at
FROM wso_rails.bulletins
WHERE type != 'Ride';
UPDATE wso_go.bulletins SET type='lostAndFound' WHERE type='LostFound';
UPDATE wso_go.bulletins SET type='job' WHERE type='Job';
UPDATE wso_go.bulletins SET type='exchange' WHERE type='Exchange';
UPDATE wso_go.bulletins SET type='announcement' WHERE type='Announcement';

INSERT INTO wso_go.bulletin_rides
(body, date, offer, source, destination, user_id, created_at, updated_at)
    SELECT body, CASE
                     WHEN start_date >= '1000-01-01 00:00:00' THEN start_date
                     ELSE created_at
    END,
    offer, source, destination, user_id, created_at, updated_at
    FROM wso_rails.bulletins
    WHERE type = 'Ride';

# courses
# Query
INSERT INTO wso_go.courses
(id, number, area_of_study_id, created_at, updated_at)
SELECT id, number, area_of_study_id, created_at, updated_at
FROM wso_rails.courses;

# departments
# Query
INSERT INTO wso_go.departments
(id, name, created_at, updated_at)
SELECT id, name, NOW(), NOW()
FROM wso_rails.departments;

# discussions
# Query
INSERT INTO wso_go.discussions
(id, title, ex_user_name, user_id, last_active, created_at, updated_at)
SELECT id, title, ex_user_name, user_id, last_active, created_at, updated_at
FROM wso_rails.discussions WHERE deleted=false;

INSERT INTO wso_go.discussions
(id, title, ex_user_name, user_id, last_active, created_at, updated_at, deleted_at)
SELECT id, title, ex_user_name, user_id, last_active, created_at, updated_at, updated_at
FROM wso_rails.discussions WHERE deleted=true;

# dorm_room
# query
INSERT INTO wso_go.dorm_rooms (
    id,
    created_at,
    updated_at,
    dorm_id,
    number,
    closet,
    flooring,
    common_room_access,
    common_room_desc,
    thermostat_access,
    key_or_card,
    faces,
    noise,
    bed_adjustable,
    hc,
    private_bathroom,
    floor_number,
    area,
    walkthrough,
    bathroom_desc,
    picture,
    room_type
)
    SELECT
    id,
    created_at,
    updated_at,
    dorm_id,
    number,
    closet,
    flooring,
    common_room_access,
    common_room_desc,
    thermostat_access,
    key_or_card,
    faces,
    noise,
    bed_adjustable,
    hc,
    private_bathroom,
    floor_number,
    area,
    walkthrough,
    bathroom_desc,
    picture,
    CASE
        WHEN room_type IS NOT NULL THEN room_type
        ELSE 'u'
        END
    FROM wso_rails.dorm_rooms;

# dorms
# query
INSERT INTO wso_go.dorms (
    id,
    created_at,
    updated_at,
    neighborhood_id,
    name,
    key_or_card,
    description,
    built,
    capacity,
    number_bathrooms,
    number_singles,
    number_doubles,
    number_washers,
    bathroom_ratio,
    loudness,
    wifi,
    location,
    satisfaction,
    average_single_area,
    average_double_area,
    mode_single_area,
    mode_double_area,
    number_flex
)
    SELECT
    id,
    created_at,
    updated_at,
    neighborhood_id,
    name,
    key_or_card,
    description,
    built,
    CASE
        WHEN capacity IS NOT NULL THEN capacity
        ELSE 0
        END,
    CASE
        WHEN number_bathrooms IS NOT NULL THEN number_bathrooms
        ELSE 0
        END,
    CASE
        WHEN number_singles IS NOT NULL THEN number_singles
        ELSE 0
        END,
    CASE
        WHEN number_doubles IS NOT NULL THEN number_doubles
        ELSE 0
        END,
    CASE
        WHEN number_washers IS NOT NULL THEN number_washers
        ELSE 0
        END,
    bathroom_ratio,
    loudness,
    wifi,
    location,
    satisfaction,
    average_single_area,
    average_double_area,
    mode_single_area,
    mode_double_area,
    CASE
        WHEN number_flex IS NOT NULL THEN number_flex
        ELSE 0
        END
FROM wso_rails.dorms;

# dormtrak_reviews
# query
INSERT INTO wso_go.dormtrak_reviews (
    id,
    created_at,
    updated_at,
    user_id,
    dorm_room_id,
    comment,
    closet,
    flooring,
    common_room_access,
    common_room_desc,
    thermostat_access,
    key_or_card,
    noise,
    bed_adjustable,
    private_bathroom,
    bathroom_desc,
    loudness,
    wifi,
    location,
    satisfaction
)
SELECT
    id,
    created_at,
    updated_at,
    user_id,
    dorm_room_id,
    comment,
    closet,
    flooring,
    common_room_access,
    common_room_desc,
    thermostat_access,
    key_or_card,
    noise,
    bed_adjustable,
    private_bathroom,
    bathroom_desc,
    loudness,
    wifi,
    location,
    satisfaction
FROM wso_rails.dormtrak_reviews;

# ephcatches
# query
INSERT INTO wso_go.ephcatches (
    id,
    created_at,
    updated_at,
    user_id,
    other_id,
    seen
)
SELECT
    id,
    created_at,
    updated_at,
    user_id,
    other_id,
    seen
FROM wso_rails.ephcatches;

# factrak agreements
# query
INSERT INTO wso_go.factrak_agreements (
    id,
    created_at,
    updated_at,
    user_id,
    factrak_survey_id,
    agrees
)
SELECT
    id,
    created_at,
    updated_at,
    user_id,
    factrak_survey_id,
    agrees
FROM wso_rails.factrak_agreements;

# factrak_surveys
# query
INSERT INTO wso_go.factrak_surveys (
    id,
    created_at,
    updated_at,
    professor_id,
    course_id,
    would_recommend_course,
    course_workload,
    course_stimulating,
    would_take_another,
    approachability,
    lead_lecture,
    promote_discussion,
    outside_helpfulness,
    user_id,
    comment,
    flagged,
    grade_received
)
SELECT
    id,
    created_at,
    updated_at,
    professor_id,
    course_id,
    would_recommend_course,
    course_workload,
    course_stimulating,
    would_take_another,
    approachability,
    lead_lecture,
    promote_discussion,
    outside_helpfulness,
    user_id,
    comment,
    flagged,
    grade_received
FROM wso_rails.factrak_surveys WHERE course_id IS NOT NULL;

# neighborhoods
# query
INSERT INTO wso_go.neighborhoods (
    id,
    created_at,
    updated_at,
    name,
    trakked
)
SELECT
    id,
    created_at,
    updated_at,
    name,
    true
FROM wso_rails.neighborhoods;
UPDATE wso_go.neighborhoods SET trakked=false WHERE name = 'First-year';
UPDATE wso_go.neighborhoods SET trakked=false WHERE name = 'Co-op';

# offices
# query
INSERT INTO wso_go.offices (
    id,
    created_at,
    updated_at,
    number
)
SELECT
    id,
    created_at,
    updated_at,
    number
FROM wso_rails.offices;

# posts
# query
INSERT INTO wso_go.posts (
    id,
    created_at,
    updated_at,
    user_id,
    discussion_id,
    content,
    ex_user_name
)
SELECT
    id,
    created_at,
    updated_at,
    user_id,
    discussion_id,
    content,
    ex_user_name
FROM wso_rails.posts WHERE deleted=false;
INSERT INTO wso_go.posts (
    id,
    created_at,
    updated_at,
    user_id,
    discussion_id,
    content,
    ex_user_name,
    deleted_at
)
    SELECT
    id,
    created_at,
    updated_at,
    user_id,
    discussion_id,
    content,
    ex_user_name,
    updated_at
FROM wso_rails.posts WHERE deleted=true;

# tags
# query
INSERT INTO wso_go.tags (
    id,
    created_at,
    updated_at,
    name
)
    SELECT
    id,
    created_at,
    updated_at,
    name
    FROM wso_rails.tags;

# users
# query
INSERT INTO wso_go.users (
    id,
    created_at,
    updated_at,
    type,
    name,
    cell_phone,
    campus_phone_ext,
    unix_id,
    williams_email,
    title,
    visible,
    class_year,
    department_id,
    dorm_visible,
    home_town,
    home_zip,
    home_state,
    home_country,
    home_visible,
    major,
    su_box,
    entry,
    admin,
    factrak_admin,
    has_accepted_factrak_policy,
    has_accepted_dormtrak_policy,
    office_id,
    dorm_room_id,
    search_fields,
    factrak_survey_deficit,
    off_cycle,
    opt_out_ephcatch,
    pronoun,
    at_williams,
    ephcatch_eligibility
)
    SELECT
    id,
    created_at,
    updated_at,
    lower(type),
    name,
    cell_phone,
    campus_phone_ext,
    unix_id,
    williams_email,
    title,
    visible,
    class_year,
    department_id,
    dorm_visible,
    home_town,
    home_zip,
    home_state,
    home_country,
    home_visible,
    major,
    su_box,
    entry,
    CASE
        WHEN admin IS NOT NULL THEN admin
        ELSE false
        END,
    factrak_admin,
    has_accepted_factrak_policy,
    has_accepted_dormtrak_policy,
    office_id,
    dorm_room_id,
    search_fields,
    factrak_survey_deficit,
    off_cycle,
    opt_out_ephcatch,
    pronoun,
    at_williams,
    ephcatch_eligibility
FROM wso_rails.users;

# tags users
# query
INSERT INTO wso_go.tags_users (
    user_id,
    tag_id
)
    SELECT
    user_id,
    tag_id
    FROM wso_rails.tags_users;