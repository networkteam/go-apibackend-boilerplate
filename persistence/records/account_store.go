package records

import "github.com/friendsofgo/errors"

func (s *AccountStore) AccountCleanupDeviceTokens(deviceToken, deviceOs string) error {
	query := "UPDATE accounts SET device_token = NULL, device_os = NULL WHERE device_token = $1 AND device_os = $2"
	_, err := s.RawExec(query, deviceToken, deviceOs)
	if err != nil {
		return errors.Wrap(err, "updating accounts to remove device token failed")
	}
	return nil
}
