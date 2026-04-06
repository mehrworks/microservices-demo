package jobcontrol

import "errors"

var ErrUnauthorized = errors.New("unauthorized")

type StaticBearerConfig struct {
	OperatorToken   string
	OperatorSubject string
	WorkerToken     string
	WorkerIdentity  WorkerIdentity
}

type StaticBearerVerifier struct {
	operatorToken   string
	operatorSubject string
	workerToken     string
	workerIdentity  WorkerIdentity
}

func NewStaticBearerVerifier(operatorToken, workerToken string) *StaticBearerVerifier {
	workerIdentity := WorkerIdentity{
		WorkerID:    "worker:static-bearer",
		AuthMode:    AuthModeBearer,
		AuthSubject: "worker:static-bearer",
	}
	return NewStaticBearerVerifierFromConfig(StaticBearerConfig{
		OperatorToken:   operatorToken,
		OperatorSubject: "operator:static-bearer",
		WorkerToken:     workerToken,
		WorkerIdentity:  workerIdentity,
	})
}

func NewStaticBearerVerifierFromConfig(cfg StaticBearerConfig) *StaticBearerVerifier {
	workerIdentity := cfg.WorkerIdentity
	if workerIdentity.AuthMode == "" {
		workerIdentity.AuthMode = AuthModeBearer
	}
	if workerIdentity.WorkerID == "" {
		workerIdentity.WorkerID = "worker:static-bearer"
	}
	if workerIdentity.AuthSubject == "" {
		workerIdentity.AuthSubject = workerIdentity.WorkerID
	}
	operatorSubject := cfg.OperatorSubject
	if operatorSubject == "" {
		operatorSubject = "operator:static-bearer"
	}

	return &StaticBearerVerifier{
		operatorToken:   cfg.OperatorToken,
		operatorSubject: operatorSubject,
		workerToken:     cfg.WorkerToken,
		workerIdentity:  workerIdentity,
	}
}

func (v *StaticBearerVerifier) VerifyOperator(rawToken string) (OperatorIdentity, error) {
	if v == nil || v.operatorToken == "" || rawToken != v.operatorToken {
		return OperatorIdentity{}, ErrUnauthorized
	}

	return OperatorIdentity{
		Subject:  v.operatorSubject,
		AuthMode: AuthModeBearer,
	}, nil
}

func (v *StaticBearerVerifier) VerifyWorker(rawToken string) (WorkerIdentity, error) {
	if v == nil || v.workerToken == "" || rawToken != v.workerToken {
		return WorkerIdentity{}, ErrUnauthorized
	}

	return v.workerIdentity, nil
}
