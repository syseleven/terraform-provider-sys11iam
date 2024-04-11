import sys
import requests

def keycloak_user1_token():
    method = "POST"
    url = "http://localhost:8181/realms/application/protocol/openid-connect/token"
    headers = {}
    params = {}
    data = {
        "grant_type": "client_credentials",
        "scope": "pytest",
        "client_id": "pytest",
        "client_secret": "YKjKvRHYtGjbxjsU2auNzcvt4FOaH5SK",
    }
    json = {}
    response = requests.request(
        method, url, headers=headers, params=params, json=json, data=data
    )
    return response.json().get("access_token")

try:
    # accept user invitation
    # get user1 membership invitations
    method = "GET"
    url = "http://localhost:9000/v1/me/invitations"
    headers = {}
    headers["Authorization"] = "Bearer {}".format(keycloak_user1_token())
    params = {}
    data = {}
    json = {}
    response = requests.request(
        method, url, headers=headers, params=params, json=json, data=data
    )

    response.raise_for_status()
    output = response.json()
    invitation_id = output[0].get("id")
except ValueError:
    print("Invitation not found.")

try:
    # accept membership by invitation
    method = "POST"
    url = "http://localhost:9000/v1/me/invitations/{0}/confirmation".format(invitation_id)
    headers = {}
    headers["Authorization"] = "Bearer {}".format(keycloak_user1_token())
    params = {}
    data = {}
    json = {}
    response = requests.request(
        method, url, headers=headers, params=params, json=json, data=data
    )
    response.raise_for_status()
except ValueError:
    print("Invitation not found.")
