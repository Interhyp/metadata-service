package acceptance

func tstUnauthenticated() string {
	return ""
}

// We created a keyset and some tokens using jwt.io. Use this keyset ONLY for automated tests!

func tstValidUserToken() string {
	return "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiZ3JvdXBzIjpbXSwiaWF0IjoxNTE2MjM5MDIyfQ.R5LZokxTi00xyVCXQlKbtkEFjGt3ezSSS9ycEspctVeTGTCwE_XPpJqbuSyJE3_U2phuAyBpIiB0qvi4_qYNsEW6xgf2eih7uqmemzQOM8jfFw_XuYPWQJ1G0LkOZMS4-q_VYgrufOUTdpECsZD9tgZOuf-zkx30UqN3-rhac3PtOtOjpv7gl5zWS8lD-eHDW7-AgrlyGWbLGfJoahGxsb-h2QOnJDDXUA4yUKYDX4Yb_9y4Np8OObCzSm0YJJlssa54L5ruluzEG_2fXvjbUDS8FlEKUqkD337ttMld41RpKnMzfXvC_mgr1zE7loDcFXQEyHSYXma5jM617RUxbg"
}

func tstValidAdminToken() string {
	return "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiZ3JvdXBzIjpbImFkbWluIl0sImlhdCI6MTUxNjIzOTAyMn0.nHGB2gk8Zyyjo5mNQm4IKF540IlbO6uR6FdzY9Oz37hGfui-pfuHAlnCoY-N1fG6xvfZe3su3EOnzOtT3BXmMEkJ3fQWexwXQKEfFDJqcAKudEvw-OTs8SO9Uc81H0d5IjGm6frJ-XKs7tSrCPGf6HC_KVx2vk19AV5eBaZNmIJTjMNBEfjgSH8lSlzpiszu-X4PYMV6j9f8PPPjdlDPMTrHx_kq1wewWhkE6TLROzdHf8w8Ip_KHRzBADdR51O_ZqONxiJMKMn7K399QpJOvEv9eI5qoYljesPwBzsFRnowNQhB3ufNWHCuCkyBM5AyW6oEo5SSkBQIry1cGUn0XA"
}

func tstInvalidSignatureToken() string {
	return "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiZ3JvdXBzIjpbImFkbWluIl0sImlhdCI6MTUxNjIzOTAyMn0.nHGB2gk8Zyyjo5mNQm4IKF540IlbO6uR6FdzY9Oz37hGfui-pfuHAlnCoY-N1fG6xvfZe3su3EOnzOtT3BXmMEkJ3fQWexwXQKEfFDJqcAKudEvw-OTs8SO9Uc81H0d5IjGm6frJ-XKs7tSrCPGf6HC_KVx2vk19AV5eBaZNmIJTjMNBEfjgSH8lSlzpiszu-X4PYMV6j9f8PPPjdlDPMTrHx_kq1wewWhkE6TLROzdHf8w8Ip_KHRzBADdR51O_ZqONxiJMKMn7K399QpJOvEv9eI5qoYljesPwBzsFRnowNQhB3ufNWHCuCkyBM5AyW6oEo5SSkBQIry1cGUn0X"
}

func tstInvalidAlgorithmToken() string {
	return "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiZ3JvdXBzIjpbImFkbWluIl0sImlhdCI6MTUxNjIzOTAyMn0.3NlwGm6CfBG5aU7myAOP14XMFtS0W5t9a9ZaS5lLLIw"
}

func tstExpiredAdminToken() string {
	return "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiZ3JvdXBzIjpbImFkbWluIl0sImlhdCI6MTUxNjIzOTAyMiwiZXhwIjoxNTE2MjM5MDQyfQ.UdKc1BvLRlyYMR2x6Nle_S5tU7QH2cZsf3jctbStR5mbou3Q9mJ0ijnFEMmrva-lAghLEKp56W65PbHz6fqDLZbe1MoiFWL6sRfvwLvSboFh2uyikuqBT7dmYvyuBpYrG6IW84-bGIodCTPZI9kNFT6x2q_nNsJxYQRv5uCe88TNAdv0JYUuRDpHGl1tXFRPievF84HfYYxrQqNz2SDIYtsCC5XXh26TsN2vNG_PBisj9UabeoumBPQcuPTASgRWTjpONxkH-8L_mzKubCVM62WFESv7ZVZ_V-DgzKkm9_r_7mVEePeBEZ-su0v2EItF0mjKW_zF-BFvvR0l417jig"
}
