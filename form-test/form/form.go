package form

import "time"

// ComputeForm represents a simple request from a control plane (admin)
// asking a data owner for minimum compute capacity information.
type ComputeForm struct {
	FormID      string `json:"form_id"`
	RequestedBy string `json:"requested_by"`
	DataOwnerID string `json:"data_owner_id,omitempty"`
	MinCompute  int    `json:"min_compute,omitempty"`
	// Additional compute and training-related fields inspired by pyproject.toml
	NumServerRounds  int               `json:"num_server_rounds,omitempty"`
	FractionEvaluate float64           `json:"fraction_evaluate,omitempty"`
	LocalEpochs      int               `json:"local_epochs,omitempty"`
	LearningRate     float64           `json:"learning_rate,omitempty"`
	BatchSize        int               `json:"batch_size,omitempty"`
	NumCPUs          int               `json:"num_cpus,omitempty"`
	NumGPUs          int               `json:"num_gpus,omitempty"`
	MemoryMB         int               `json:"memory_mb,omitempty"`
	Model            string            `json:"model,omitempty"`
	Framework        string            `json:"framework,omitempty"`
	Components       map[string]string `json:"components,omitempty"`

	// Role tracking
	DataOwnerFilled   bool `json:"data_owner_filled,omitempty"`
	OutputOwnerFilled bool `json:"output_owner_filled,omitempty"`
	// Overall filled flag (true when both roles have filled their parts)
	Filled      bool       `json:"filled"`
	RequestedAt time.Time  `json:"requested_at"`
	FilledAt    *time.Time `json:"filled_at,omitempty"`
}

// NewEmptyComputeForm creates an empty form requested by an admin/control plane.
func NewEmptyComputeForm(formID, requestedBy string) ComputeForm {
	return ComputeForm{
		FormID:      formID,
		RequestedBy: requestedBy,
		Filled:      false,
		RequestedAt: time.Now().UTC(),
	}
}

// ComputeSpec groups optional compute and training configuration values.
type ComputeSpec struct {
	NumServerRounds  int
	FractionEvaluate float64
	LocalEpochs      int
	LearningRate     float64
	BatchSize        int
	NumCPUs          int
	NumGPUs          int
	MemoryMB         int
	Model            string
	Framework        string
	Components       map[string]string
}

// FillByDataOwner fills only the fields a data owner should provide.
func FillByDataOwner(f *ComputeForm, ownerID string, minCompute, numCPUs, numGPUs, memoryMB int) {
	f.DataOwnerID = ownerID
	f.MinCompute = minCompute
	f.NumCPUs = numCPUs
	f.NumGPUs = numGPUs
	f.MemoryMB = memoryMB
	f.DataOwnerFilled = true

	// update overall filled flag and timestamp if both participants have filled
	f.Filled = f.DataOwnerFilled && f.OutputOwnerFilled
	if f.Filled {
		t := time.Now().UTC()
		f.FilledAt = &t
	}
}

// FillByOutputOwner fills only the fields the output owner should provide.
func FillByOutputOwner(f *ComputeForm, ownerID string, spec ComputeSpec) {
	// apply spec values
	f.NumServerRounds = spec.NumServerRounds
	f.FractionEvaluate = spec.FractionEvaluate
	f.LocalEpochs = spec.LocalEpochs
	f.LearningRate = spec.LearningRate
	f.BatchSize = spec.BatchSize
	f.Framework = spec.Framework
	if spec.Components != nil {
		f.Components = spec.Components
	}

	f.OutputOwnerFilled = true
	f.Filled = f.DataOwnerFilled && f.OutputOwnerFilled
	if f.Filled {
		t := time.Now().UTC()
		f.FilledAt = &t
	}
}

// FillComputeFormWithSpec is a backwards-compatible helper that fills both sides.
func FillComputeFormWithSpec(f *ComputeForm, ownerID string, minCompute int, spec ComputeSpec) {
	// fill data-owner parts
	FillByDataOwner(f, ownerID, minCompute, spec.NumCPUs, spec.NumGPUs, spec.MemoryMB)
	// fill output-owner parts
	FillByOutputOwner(f, ownerID, spec)
}

// --- New: two separate forms ---

// DataOwnerForm contains only data-owner responsibilities (resources and data identifiers).
type DataOwnerForm struct {
	FormID         string     `json:"form_id"`
	RequestedBy    string     `json:"requested_by"`
	DataOwnerID    string     `json:"data_owner_id,omitempty"`
	MinCompute     int        `json:"min_compute,omitempty"`
	NumCPUs        int        `json:"num_cpus,omitempty"`
	NumGPUs        int        `json:"num_gpus,omitempty"`
	MemoryMB       int        `json:"memory_mb,omitempty"`
	DataSizeBytes  int64      `json:"data_size_bytes,omitempty"`
	DataResourceID string     `json:"data_resource_id,omitempty"`
	Filled         bool       `json:"filled"`
	RequestedAt    time.Time  `json:"requested_at"`
	FilledAt       *time.Time `json:"filled_at,omitempty"`
}

// OutputOwnerForm contains only output-owner responsibilities (training params, metadata).
type OutputOwnerForm struct {
	FormID           string            `json:"form_id"`
	RequestedBy      string            `json:"requested_by"`
	OutputOwnerID    string            `json:"output_owner_id,omitempty"`
	NumServerRounds  int               `json:"num_server_rounds,omitempty"`
	FractionEvaluate float64           `json:"fraction_evaluate,omitempty"`
	LocalEpochs      int               `json:"local_epochs,omitempty"`
	LearningRate     float64           `json:"learning_rate,omitempty"`
	BatchSize        int               `json:"batch_size,omitempty"`
	Model            string            `json:"model,omitempty"`
	Framework        string            `json:"framework,omitempty"`
	Components       map[string]string `json:"components,omitempty"`
	Filled           bool              `json:"filled"`
	RequestedAt      time.Time         `json:"requested_at"`
	FilledAt         *time.Time        `json:"filled_at,omitempty"`
}

// NewEmptyDataOwnerForm creates an empty DataOwnerForm.
func NewEmptyDataOwnerForm(formID, requestedBy string) DataOwnerForm {
	return DataOwnerForm{FormID: formID, RequestedBy: requestedBy, RequestedAt: time.Now().UTC()}
}

// NewEmptyOutputOwnerForm creates an empty OutputOwnerForm.
func NewEmptyOutputOwnerForm(formID, requestedBy string) OutputOwnerForm {
	return OutputOwnerForm{FormID: formID, RequestedBy: requestedBy, RequestedAt: time.Now().UTC()}
}

// FillDataOwnerForm fills fields the data owner should provide.
func FillDataOwnerForm(f *DataOwnerForm, ownerID string, minCompute, numCPUs, numGPUs, memoryMB int, dataSizeBytes int64, dataResourceID string) {
	f.DataOwnerID = ownerID
	f.MinCompute = minCompute
	f.NumCPUs = numCPUs
	f.NumGPUs = numGPUs
	f.MemoryMB = memoryMB
	f.DataSizeBytes = dataSizeBytes
	f.DataResourceID = dataResourceID
	f.Filled = true
	t := time.Now().UTC()
	f.FilledAt = &t
}

// FillOutputOwnerForm fills fields the output owner should provide.
func FillOutputOwnerForm(f *OutputOwnerForm, ownerID string, spec ComputeSpec) {
	f.OutputOwnerID = ownerID
	f.NumServerRounds = spec.NumServerRounds
	f.FractionEvaluate = spec.FractionEvaluate
	f.LocalEpochs = spec.LocalEpochs
	f.LearningRate = spec.LearningRate
	f.BatchSize = spec.BatchSize
	f.Model = spec.Model
	f.Framework = spec.Framework
	if spec.Components != nil {
		f.Components = spec.Components
	}
	f.Filled = true
	t := time.Now().UTC()
	f.FilledAt = &t
}
