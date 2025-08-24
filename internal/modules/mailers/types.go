package mailers

type StrategyNames struct{}

// #region Auth strategies

func (StrategyNames) Verification() string  { return "verification" }
func (StrategyNames) PasswordReset() string { return "password_reset" }
func (StrategyNames) Welcome() string       { return "welcome" }

// #endregion

// Strategies provides easy access to strategy names
var Strategies = StrategyNames{}
