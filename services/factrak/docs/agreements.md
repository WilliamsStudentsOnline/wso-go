# Agreements

Controller for all agreement methods.

## Get Agreement

Gets a user's agreement for a specific survey.

**URL**: `GET /surveys/:surveyID/agreement`

**JWT required**: YES

**Scopes required**: `service:factrak:full`

## Success Response

**Code** : `200 OK`

**Content examples**

```json
{
  "agrees": true,
  "factrakSurveyID": 10,
  "userID": 42
}
```

## Notes