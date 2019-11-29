// Handlers perform actions on persistence and services through domain commands
//
// API or CLI or other endpoints will call handlers to cause side-effects (write).
// Each handler should perform a transactional step. It should either succeed or fail completely.
package handler
