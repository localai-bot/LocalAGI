package types

type JobFilter interface {
	Name() string
	Apply(job *Job) (bool, error)
}

type JobFilters []JobFilter
