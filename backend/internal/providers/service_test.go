package providers

import (
	"context"
	"reflect"
	"testing"

	"github.com/eaguilar88/deu/internal/entities"
	"go.uber.org/zap"
)

func TestNewProvidersService(t *testing.T) {
	type args struct {
		repo    Repository
		storage StorageClient
		logger  *zap.Logger
	}
	tests := []struct {
		name string
		args args
		want *ProvidersService
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewProvidersService(tt.args.repo, tt.args.storage, tt.args.logger); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewProvidersService() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProvidersService_GetProvider(t *testing.T) {
	type fields struct {
		repo    Repository
		storage StorageClient
		logger  *zap.Logger
	}
	type args struct {
		ctx        context.Context
		providerID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    entities.Provider
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ProvidersService{
				repo:    tt.fields.repo,
				storage: tt.fields.storage,
				logger:  tt.fields.logger,
			}
			got, err := s.GetProvider(tt.args.ctx, tt.args.providerID)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProvidersService.GetProvider() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ProvidersService.GetProvider() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProvidersService_GetProviderByCode(t *testing.T) {
	type fields struct {
		repo    Repository
		storage StorageClient
		logger  *zap.Logger
	}
	type args struct {
		ctx  context.Context
		code string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    entities.Provider
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ProvidersService{
				repo:    tt.fields.repo,
				storage: tt.fields.storage,
				logger:  tt.fields.logger,
			}
			got, err := s.GetProviderByCode(tt.args.ctx, tt.args.code)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProvidersService.GetProviderByCode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ProvidersService.GetProviderByCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProvidersService_GetProviders(t *testing.T) {
	type fields struct {
		repo    Repository
		storage StorageClient
		logger  *zap.Logger
	}
	type args struct {
		ctx       context.Context
		pageScope entities.PageScope
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []entities.Provider
		want1   entities.PageScope
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ProvidersService{
				repo:    tt.fields.repo,
				storage: tt.fields.storage,
				logger:  tt.fields.logger,
			}
			got, got1, err := s.GetProviders(tt.args.ctx, tt.args.pageScope)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProvidersService.GetProviders() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ProvidersService.GetProviders() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("ProvidersService.GetProviders() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestProvidersService_CreateProvider(t *testing.T) {
	type fields struct {
		repo    Repository
		storage StorageClient
		logger  *zap.Logger
	}
	type args struct {
		ctx      context.Context
		provider *entities.Provider
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int64
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ProvidersService{
				repo:    tt.fields.repo,
				storage: tt.fields.storage,
				logger:  tt.fields.logger,
			}
			got, _, err := s.CreateProvider(tt.args.ctx, tt.args.provider)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProvidersService.CreateProvider() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ProvidersService.CreateProvider() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProvidersService_UpdateProvider(t *testing.T) {
	type fields struct {
		repo    Repository
		storage StorageClient
		logger  *zap.Logger
	}
	type args struct {
		ctx        context.Context
		providerID string
		provider   *entities.Provider
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ProvidersService{
				repo:    tt.fields.repo,
				storage: tt.fields.storage,
				logger:  tt.fields.logger,
			}
			if err := s.UpdateProvider(tt.args.ctx, tt.args.providerID, tt.args.provider); (err != nil) != tt.wantErr {
				t.Errorf("ProvidersService.UpdateProvider() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestProvidersService_DeleteProvider(t *testing.T) {
	type fields struct {
		repo    Repository
		storage StorageClient
		logger  *zap.Logger
	}
	type args struct {
		ctx        context.Context
		providerID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ProvidersService{
				repo:    tt.fields.repo,
				storage: tt.fields.storage,
				logger:  tt.fields.logger,
			}
			if err := s.DeleteProvider(tt.args.ctx, tt.args.providerID); (err != nil) != tt.wantErr {
				t.Errorf("ProvidersService.DeleteProvider() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestProvidersService_uploadAndSave(t *testing.T) {
	type fields struct {
		repo    Repository
		storage StorageClient
		logger  *zap.Logger
	}
	type args struct {
		ctx   context.Context
		files []*entities.File
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ProvidersService{
				repo:    tt.fields.repo,
				storage: tt.fields.storage,
				logger:  tt.fields.logger,
			}
			if err := s.storage.UploadFile(tt.args.ctx, tt.args.files); (err != nil) != tt.wantErr {
				t.Errorf("ProvidersService.uploadAndSave() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_prepareFilesSlice(t *testing.T) {
	type args struct {
		provider   *entities.Provider
		providerID int64
		metadata   map[string]string
	}
	tests := []struct {
		name    string
		args    args
		want    []*entities.File
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := prepareFilesSlice(tt.args.provider, tt.args.providerID, tt.args.metadata)
			if (err != nil) != tt.wantErr {
				t.Errorf("prepareFilesSlice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("prepareFilesSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_makeFileEntityFromFilePointer(t *testing.T) {
	type args struct {
		file       *entities.File
		providerID int64
		uploadedBy string
		ownerType  entities.OwnerType
		metadata   map[string]string
	}
	tests := []struct {
		name string
		args args
		want *entities.File
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := makeFileEntityFromFilePointer(tt.args.file, tt.args.providerID, tt.args.uploadedBy, tt.args.ownerType, tt.args.metadata); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("makeFileEntityFromFilePointer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProvidersService_getFilesForProvider(t *testing.T) {
	type fields struct {
		repo    Repository
		storage StorageClient
		logger  *zap.Logger
	}
	type args struct {
		ctx        context.Context
		providerID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    entities.ProviderFiles
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ProvidersService{
				repo:    tt.fields.repo,
				storage: tt.fields.storage,
				logger:  tt.fields.logger,
			}
			got, err := s.getFilesForProvider(tt.args.ctx, tt.args.providerID)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProvidersService.getFilesForProvider() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ProvidersService.getFilesForProvider() = %v, want %v", got, tt.want)
			}
		})
	}
}
