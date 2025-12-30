package command

type AuthSessionLogoutCmd struct{}

func NewAuthSessionLogoutCmd() AuthSessionLogoutCmd {
	return AuthSessionLogoutCmd{}
}

func (c AuthSessionLogoutCmd) Validate() error {
	return nil
}
