# List Users

Lists all users that are visible and at Williams.

**URL**: `/users`

**Method**: `GET`

**JWT required**: YES

**Scopes required**: `service:users`

## Success Response

**Code** : `200 OK`

**Content examples**

```json
[
    {
        "id": 1,
        "type": "Student",
        "name": "Zane Rogahn",
        "cell_phone": "1-619-417-6008",
        "campus_phone_ext": "1311",
        "unix_id": "student",
        "williams_email": "student@williams.edu",
        "title": "Dynamic Engineer",
        "visible": false,
        "class_year": 2021,
        "department_id": 33,
        "dorm_visible": true,
        "home_town": "North Wallyview",
        "home_zip": "13555-6198",
        "home_phone": "(490) 122-5892 x140",
        "home_state": "Connecticut",
        "home_country": "Bhutan",
        "home_visible": true,
        "major": "PSCI",
        "su_box": "3330",
        "entry": "Sage",
        "admin": false,
        "factrak_admin": false,
        "has_accepted_factrak_policy": true,
        "has_accepted_dormtrak_policy": false,
        "pronoun": "",
        "at_williams": true,
        "off_cycle": false,
        "factrak_survey_deficit": 2,
        "opt_out_ephcatch": false,
        "ephcatch_eligibility": false
    }
]
```

## Notes