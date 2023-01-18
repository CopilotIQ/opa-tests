# JWT Authorization Policy
# Created M. Massenzio, 2021-11-15
#
# This should be loaded to the OPA Policy Server via a PUT request to the /v1/policies endpoint.

package copilotiq
import data.copilotiq.common as c
import future.keywords.in

default allow = false

# System accounts are allowed to make all API calls.
allow {
  c.is_system
}

# User is allowed to view/modify self but cannot create/delete itself,
# neither execute extra actions (roles, status, username)
allow {
    c.entity == "users"
    c.is_user
    c.entity_id == c.user
    not c.extra
    input.resource.method in [ "GET", "PUT"]
}

# Admin is allowed to view/create/delete all users.
allow {
    c.entity == "users"
    c.is_admin
    input.resource.method in ["GET", "DELETE", "POST"]
}

# Admin is allowed to update users status.
allow {
  c.is_admin
  c.entity == "users"
  c.extra == "status"
  input.resource.method == "PUT"
}

# Admin is allowed to update users roles.
# See ENG-370
allow {
  c.is_admin
  c.entity == "users"
  c.extra == "roles"
  input.resource.method == "PUT"
}

# Leadership is allowed to view users.
# See ENG-652
allow {
  c.is_leadership
  c.entity == "users"
  input.resource.method == "GET"
}

# ENG-215: Medical staff are allowed to view patients details.
allow {
  c.is_medical_staff
  c.entity == "users"
  input.resource.method == "GET"
}

# ENG-352: Medical staff are allowed to view patients accounts.
allow {
  c.is_medical_staff
  c.entity == "accounts"
  input.resource.method == "GET"
}

# ENG-352: NPS can create accounts
allow {
  c.is_nps
  c.entity == "accounts"
  input.resource.method == "POST"
}

# ENG-352: NPS can add/update accounts
allow {
  c.is_nps
  c.entity == "users"
  c.extra == "accounts"
  input.resource.method == "PUT"
}

# Admin is allowed all operations on Accounts.
allow {
    c.is_admin
    c.entity == "accounts"
}
