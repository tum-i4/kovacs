import sys
from utils import default_url_base, make_post_request


def create_user(email: str, password: str, first_name: str, last_name: str, url: str) -> bool:
	# Clean up input
	email = email.strip()
	password = password.strip()
	first_name = first_name.strip()
	last_name = last_name.strip()

	if len(first_name) == 0:
		first_name = 'First'

	if len(last_name) == 0:
		last_name = 'Last'

	# Create request
	request_data = {
		'email': email,
		'password': password,
		'firstName': first_name,
		'lastName': last_name
	}

	resp, ok = make_post_request(url, request_data, debug=False, print_success=False)
	if ok:
		return True

	# User with given ID already exists
	if resp.status_code == 409 and resp.text == "{\"error\":{\"message\":\"user with given ID already exists\"}}":
		return True

	print(f'! Failure ({resp.status_code}): {resp.text}')
	return False


userAmount = 2
url_base = default_url_base
if len(sys.argv) > 3:
	print("Invalid parameter amount")
	print("Must have a maximum of two parameters")
	print(f"First parameter is the amount of users to be created. Defaults to {userAmount}")
	print(f"Second parameter is Revolori's URL. Defaults to '{url_base}'")
	sys.exit(1)

if len(sys.argv) >= 2:
	try:
		userAmount = int(sys.argv[1])
	except ValueError:
		print(f"Got an invalid number: '{sys.argv[1]}'\nExiting")
		sys.exit(1)

if len(sys.argv) == 3:
	url_base = sys.argv[2]

exit_code = 0
for i in range(0, userAmount):
	try:
		if not create_user(
			email=f"user{i}@example.com",
			password='password',
			first_name=f"First {i}",
			last_name=f"Last {i}",
			url=url_base + "/user"
		):
			exit_code = 1
	except Exception as e:
		print(f"An error occurred: {e}")
		sys.exit(1)

exit(exit_code)
