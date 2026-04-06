package jobcontrol

import "errors"

var ErrUnauthorized = errors.New("unauthorized")

type StaticBearerVerifier struct {
	operatorToken string
	workerToken   string
}

func NewStaticBearerVerifier(operatorToken, workerToken string) *StaticBearerVerifier {
	return &StaticBearerVerifier{
		operatorToken: operatorToken,
		workerToken:   workerToken,
	}
}

func (v *StaticBearerVerifier) VerifyOperator(rawToken string) (OperatorIdentity, error) {
	if v == nil || v.operatorToken == "" || rawToken != v.operatorToken {
		return OperatorIdentity{}, ErrUnauthorized
	}

	return OperatorIdentity{
		Subject:  "operator:static-bearer",
		AuthMode: AuthModeBearer,
	}, nil
}

func (v *StaticBearerVerifier) VerifyWorker(rawToken string) (WorkerIdentity, error) {
	if v == nil || v.workerToken == "" || rawToken != v.workerToken {
		return WorkerIdentity{}, ErrUnauthorized
	}

	return WorkerIdentity{
		WorkerID:    "worker:static-bearer",
		AuthMode:    AuthModeBearer,
		AuthSubject: "worker:static-bearer",
	}, nil
}
