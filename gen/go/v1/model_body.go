/*
 * Vendor API V1
 *
 * Apps documentation
 *
 * API version: 1.0.0
 * Contact: info@replicated.com
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */

package swagger

type Body struct {
	// Name of the app that is to be created.
	Name string `json:"name"`
	// Scheduler of the app that is to be created
	Scheduler string `json:"scheduler,omitempty"`
}
