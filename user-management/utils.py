import json
from typing import Optional
from requests.auth import HTTPBasicAuth
import requests

default_url_base = 'http://localhost:5429'


def make_post_request(url: str, request_data: dict, debug=False, username='admin', print_success=True) -> Optional[requests.Response]:
	request_data_str = json.dumps(request_data)

	if debug:
		print(f'URL: {url}\nRequest: {request_data_str}')
		print('=====')

	resp = requests.post(url, request_data_str, auth=HTTPBasicAuth(username, 'password'))

	if resp.status_code == 200 or resp.status_code == 201:
		if print_success:
			print(f'Success: {resp.text}')

		return resp

	if username == 'admin':
		return make_post_request(url=url, request_data=request_data, debug=debug, username="user")

	print(f'! Failure ({resp.status_code}): {resp.text}')
	return None
