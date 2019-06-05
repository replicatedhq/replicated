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

// An app channel belongs to an app. It contains references to the top (current) release in the channel.
type AppChannel struct {
	Adoption *ChannelAdoption `json:"Adoption,omitempty"`
	// AirgapEnabled indicates if airgap builds will be created automatically for this channel
	AirgapEnabled bool `json:"AirgapEnabled,omitempty"`
	// Description that will be shown during license installation
	Description string `json:"Description"`
	// The ID of the channel
	Id string `json:"Id"`
	// IsDefault shows if channel is default or not
	IsDefault     bool           `json:"IsDefault,omitempty"`
	LicenseCounts *LicenseCounts `json:"LicenseCounts,omitempty"`
	// The name of channel
	Name string `json:"Name"`
	// The position for which the channel occurs in a list
	Position int64 `json:"Position,omitempty"`
	// The label of the current release sequence
	ReleaseLabel string `json:"ReleaseLabel,omitempty"`
	// Release notes for the current release sequence
	ReleaseNotes string `json:"ReleaseNotes,omitempty"`
	// A reference to the current release sequence
	ReleaseSequence int64 `json:"ReleaseSequence,omitempty"`
}
