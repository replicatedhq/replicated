package main

import (
	"bytes"
	"context"
	"dagger/replicated/internal/dagger"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type Logs struct {
	Stdout string `json:"stdout"`
	Stderr string `json:"stderr"`
}

type ValidationResult struct {
	IsSecurityPass bool            `json:"isSecurityPass"`
	SecurityLogs   map[string]Logs `json:"securityLogs"`

	IsFunctionalityPass bool            `json:"isFunctionalityPass"`
	FunctionalityLogs   map[string]Logs `json:"functionalityLogs"`

	IsCompatibilityPass bool            `json:"isCompatibilityPass"`
	CompatibilityLogs   map[string]Logs `json:"compatibilityLogs"`

	IsPerformancePass bool            `json:"isPerformancePass"`
	PerformanceLogs   map[string]Logs `json:"performanceLogs"`
}

func (r *Replicated) Validate(
	ctx context.Context,

	// +defaultPath="./"
	source *dagger.Directory,

	// +optional
	onePasswordServiceAccount *dagger.Secret,
) error {
	validationResult := ValidationResult{}

	securityOK, securityLogs, err := validateSecurity(ctx, source)
	if err != nil {
		return err
	}
	validationResult.IsSecurityPass = securityOK
	validationResult.SecurityLogs = securityLogs

	functionalityOK, functionalityLogs, err := validateFunctionality(ctx, source)
	if err != nil {
		return err
	}
	validationResult.IsFunctionalityPass = functionalityOK
	validationResult.FunctionalityLogs = functionalityLogs

	compatibilityOK, copmfunctionalityLogs, err := validateCompatibility(ctx, source)
	if err != nil {
		return err
	}
	validationResult.IsCompatibilityPass = compatibilityOK
	validationResult.CompatibilityLogs = copmfunctionalityLogs

	performanceOK, performanceLogs, err := validatePerformance(ctx, source)
	if err != nil {
		return err
	}
	validationResult.IsPerformancePass = performanceOK
	validationResult.PerformanceLogs = performanceLogs

	if onePasswordServiceAccount != nil {
		container := dag.Container().
			From("alpine/git:latest").
			WithMountedDirectory("/go/src/github.com/replicatedhq/replicated", source).
			WithWorkdir("/go/src/github.com/replicatedhq/replicated").
			With(CacheBustingExec([]string{"git", "status", "--porcelain"}))

		gitStatusOutput, err := container.Stdout(ctx)
		if err != nil {
			return err
		}

		gitTreeClean := len(strings.TrimSpace(gitStatusOutput)) == 0

		container = dag.Container().
			From("alpine/git:latest").
			WithMountedDirectory("/go/src/github.com/replicatedhq/replicated", source).
			WithWorkdir("/go/src/github.com/replicatedhq/replicated").
			With(CacheBustingExec([]string{"git", "rev-parse", "HEAD"}))

		commit, err := container.Stdout(ctx)
		if err != nil {
			return err
		}
		commit = strings.TrimSpace(commit)

		if gitTreeClean {
			if err := uploadValidationResult(ctx, onePasswordServiceAccount, validationResult, commit); err != nil {
				return err
			}
		}
	}

	return nil
}

func uploadValidationResult(ctx context.Context, onePasswordServiceAccount *dagger.Secret, validationResult ValidationResult, commit string) error {
	b, err := json.Marshal(validationResult)
	if err != nil {
		return err
	}

	// i think i need to create a new container to write b to a dagger.File?
	f := dag.Directory().
		WithNewFile("validation-result.json", string(b)).
		File("validation-result.json")
	f = f.WithName(commit)
	accessKeyID, err := dag.Onepassword().FindSecret(
		onePasswordServiceAccount,
		"Developer Automation",
		"S3 Workflow Validation",
		"access_key_id").Plaintext(ctx)
	if err != nil {
		return err
	}

	secretAccessKey, err := dag.Onepassword().FindSecret(
		onePasswordServiceAccount,
		"Developer Automation",
		"S3 Workflow Validation",
		"secret_access_key").Plaintext(ctx)
	if err != nil {
		return err
	}

	sess := session.Must(session.NewSession(&aws.Config{
		Region:      aws.String("us-east-1"),
		Credentials: credentials.NewStaticCredentials(accessKeyID, secretAccessKey, ""),
	}))

	svc := s3.New(sess)
	f.Export(ctx, commit)

	readFile, err := os.ReadFile(commit)
	if err != nil {
		return err
	}

	_, err = svc.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket:             aws.String("workflow-validation-results"),
		Body:               bytes.NewReader(readFile),
		Key:                aws.String(commit),
		ContentDisposition: aws.String("attachment"),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == request.CanceledErrorCode {
			fmt.Fprintf(os.Stderr, "upload canceled due to timeout, %v\n", err)
			return err
		} else {
			fmt.Fprintf(os.Stderr, "failed to upload object, %v\n", err)
			return err
		}
	}

	return nil
}
