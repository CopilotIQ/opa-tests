# JWT Authorization Policy
# Created M. Massenzio, 2021-11-15
#
# Common functionality


package copilotiq.common
import future.keywords.in

# The JWT carries the username and roles, which will be used
# to authorize access to the endpoint (`input.resource.path`)
token := t[1] {
    t := io.jwt.decode(input.api_token)
}

user := u {
    token.iss == "copilotiq.com"
    u = token.sub
}

roles := r {
    r = token.roles
}

# SYSTEM roles (typically, only bots) are allowed to make any
# API calls, with whatever HTTP Method.
is_system {
    some i, "SYSTEM" in roles
}

# Admin users can only create/modify a subset
# of entities, but is still a powerful role, ASSIGN WITH CARE.
is_admin {
    some i, "ADMIN" in roles
}

# Users can only modify self, and entities associated
# with the users themselves.
# We assume that the user is valid if it could obtain a valid JWT and
# has at least one Role.
is_user {
    count(roles) > 0
}

is_manager {
    some i
      endswith(roles[i], "MANAGER")
}

# Simplified method to split the Path into a mapping {entity, id, extra, query_args}
# TODO: Further parse the Query args into a mapping.
split_path(path) = s {
    t := trim(path, "/")
    q := split(t, "?")
    s := split(q[0], "/")
}

# Split the path segments into their constituents
segments = split_path(input.resource.path)
entity := segments[0]
entity_id := segments[1]
extra := segments[2]
