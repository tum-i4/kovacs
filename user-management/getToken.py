import sys
from utils import default_url_base, make_post_request


def login_user(email: str, password: str, url: str) -> bool:
	# Clean up input
	email = email.strip()
	password = password.strip()

	# Create request
	request_data = {
		'email': email,
		'password': password,
	}

	response_data, ok = make_post_request(url, request_data, debug=False, print_success=False)
	if not ok:
		return False

	try:
		token = response_data.cookies.get("token")
	except:
		return False

	if len(token) == 0:
		return False

	print(token)
	return True


url_base = default_url_base
if len(sys.argv) != 2 and len(sys.argv) != 3:
	print("Invalid parameter amount")
	print("Must have at least one parameter and a maximum of two")
	print("First parameter is id of users you wish to get the token for")
	print(f"Second parameter is Revolori's URL. Defaults to '{url_base}'")
	sys.exit(1)

userID = 0
try:
	userID = int(sys.argv[1])
except ValueError:
	print(f"Got an invalid number: '{sys.argv[1]}'\nExiting")
	sys.exit(1)

if len(sys.argv) == 3:
	url_base = sys.argv[2]

exit(
	not login_user(
		email=f"user{userID}@example.com",
		password="password",
		url=url_base + '/login'
	)
)
