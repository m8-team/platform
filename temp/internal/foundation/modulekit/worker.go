package modulekit

type WorkerRegistration struct {
	Name     string
	Provider any
}

type WorkerRegistrar interface {
	RegisterWorker(worker WorkerRegistration)
}

type WorkerRegistry struct {
	workers []WorkerRegistration
}

func NewWorkerRegistry() *WorkerRegistry {
	return &WorkerRegistry{workers: make([]WorkerRegistration, 0)}
}

func (r *WorkerRegistry) RegisterWorker(worker WorkerRegistration) {
	if r == nil {
		return
	}
	r.workers = append(r.workers, worker)
}

func (r *WorkerRegistry) Workers() []WorkerRegistration {
	if r == nil {
		return nil
	}
	workers := make([]WorkerRegistration, len(r.workers))
	copy(workers, r.workers)
	return workers
}
