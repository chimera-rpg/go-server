package GameWorld

// Interface gameOwner represents the general interface that should be used
// for controlling and managing autonomous gameObject(s). It is used for
// Players and will eventually be used for NPCs.
type gameOwner interface {
  getTarget() *GameObject
  setTarget(*GameObject)
}
