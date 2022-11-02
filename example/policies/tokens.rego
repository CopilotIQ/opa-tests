# Authorization Policy for API Token Generation and Refresh
# Created M. Massenzio, 2022-06-07

package copilotiq
import data.copilotiq.common as c

# Users are allowed to request their tokens to be refreshed.
allow {
    c.is_user
    input.resource.method == "GET"
    c.entity = "token"
}
