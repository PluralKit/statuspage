#!/bin/sh

url=http://localhost:8080

incidentA=$(cat << END
{
    "status": "identified",
    "impact": "minor",
    "name": "Avatars not loading",
    "description": "Some avatars are not loading."
}
END
)
incidentA_updateA=$(cat << END
{
    "text": "uhhhh there's a fox in the servers biting on wires???"
}
END
)
incidentA_updateB=$(cat << END
{
    "text": "we have lured the fox out with a sandwich! working on repairing the wires now"
}
END
)

incidentB=$(cat << END
{
    "status": "investigating",
    "impact": "major",
    "name": "Myriad fell asleep :(",
    "description": "PluralKit currently isn't working because Myriad is taking a nap."
}
END
)

incidentA_ID=$(curl --header "Content-Type: application/json" \
  --silent \
  --request POST \
  --data "$incidentA" \
  "$url/api/v1/admin/incidents/create"
)

curl --header "Content-Type: application/json" \
  --silent \
  --request POST \
  --data "$incidentB" \
  "$url/api/v1/admin/incidents/create"

curl --header "Content-Type: application/json" \
  --silent \
  --request POST \
  --data "$incidentA_updateA" \
  "$url/api/v1/admin/incidents/$incidentA_ID/update"

curl --header "Content-Type: application/json" \
  --silent \
  --request POST \
  --data "$incidentA_updateB" \
  "$url/api/v1/admin/incidents/$incidentA_ID/update"