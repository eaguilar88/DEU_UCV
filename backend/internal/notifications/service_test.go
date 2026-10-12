package notifications

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/eaguilar88/deu/internal/email"
	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/notifications/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

var (
	testNow    = time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)
	testConfig = Config{
		PendingRequestAge: 72 * time.Hour,
		RenewalWindow:     30 * 24 * time.Hour,
		ResendInterval:    7 * 24 * time.Hour,
	}
	resendSince = testNow.Add(-testConfig.ResendInterval)
)

func newTestService(repo Repository, mail MailClient) *Service {
	svc := NewService(repo, mail, testConfig, zap.NewNop())
	svc.now = func() time.Time { return testNow }
	return svc
}

func TestService_SendCoordinatorReminders(t *testing.T) {
	createdBefore := testNow.Add(-testConfig.PendingRequestAge)
	kind := entities.NotificationKindCoordinatorPendingRequests

	tests := []struct {
		name    string
		prepare func(repo *mocks.MockRepository, mail *mocks.MockMailClient)
		wantErr bool
	}{
		{
			name: "sends one summary per coordinator with course and group counts",
			prepare: func(repo *mocks.MockRepository, mail *mocks.MockMailClient) {
				repo.EXPECT().CountStaleCourseRequestsByFaculty(mock.Anything, createdBefore).
					Return([]entities.FacultyPendingCount{{Faculty: entities.FacultyCiencias, Count: 2}}, nil)
				repo.EXPECT().CountStaleGroupRequestsByFaculty(mock.Anything, createdBefore).
					Return([]entities.FacultyPendingCount{{Faculty: entities.FacultyCiencias, Count: 1}}, nil)
				repo.EXPECT().GetFacultyCoordinatorEmails(mock.Anything, entities.FacultyCiencias).
					Return([]string{"coord@example.com"}, nil)
				repo.EXPECT().WasNotifiedSince(mock.Anything, kind, "Ciencias", "coord@example.com", resendSince).Return(false, nil)
				mail.EXPECT().SendTemplate(mock.Anything, "coord@example.com", email.TemplateCoordinatorPendingReminder,
					email.CoordinatorPendingReminderData{Faculty: "Ciencias", MinDays: 3, CourseRequests: 2, GroupRequests: 1}).Return(nil)
				repo.EXPECT().LogNotification(mock.Anything, entities.NotificationLog{Kind: kind, Target: "Ciencias", Recipient: "coord@example.com"}).Return(nil)
			},
		},
		{
			name: "DEU requests also go to the DEU admins without duplicates",
			prepare: func(repo *mocks.MockRepository, mail *mocks.MockMailClient) {
				repo.EXPECT().CountStaleCourseRequestsByFaculty(mock.Anything, createdBefore).
					Return([]entities.FacultyPendingCount{{Faculty: entities.FacultyDEU, Count: 1}}, nil)
				repo.EXPECT().CountStaleGroupRequestsByFaculty(mock.Anything, createdBefore).Return(nil, nil)
				repo.EXPECT().GetFacultyCoordinatorEmails(mock.Anything, entities.FacultyDEU).Return([]string{"a@example.com"}, nil)
				repo.EXPECT().GetDEUAdminEmails(mock.Anything).Return([]string{"a@example.com", "b@example.com"}, nil)
				for _, to := range []string{"a@example.com", "b@example.com"} {
					repo.EXPECT().WasNotifiedSince(mock.Anything, kind, "DEU", to, resendSince).Return(false, nil).Once()
					mail.EXPECT().SendTemplate(mock.Anything, to, email.TemplateCoordinatorPendingReminder, mock.Anything).Return(nil).Once()
					repo.EXPECT().LogNotification(mock.Anything, entities.NotificationLog{Kind: kind, Target: "DEU", Recipient: to}).Return(nil).Once()
				}
			},
		},
		{
			name: "skips a coordinator reminded within the resend interval",
			prepare: func(repo *mocks.MockRepository, mail *mocks.MockMailClient) {
				repo.EXPECT().CountStaleCourseRequestsByFaculty(mock.Anything, createdBefore).
					Return([]entities.FacultyPendingCount{{Faculty: entities.FacultyCiencias, Count: 2}}, nil)
				repo.EXPECT().CountStaleGroupRequestsByFaculty(mock.Anything, createdBefore).Return(nil, nil)
				repo.EXPECT().GetFacultyCoordinatorEmails(mock.Anything, entities.FacultyCiencias).Return([]string{"coord@example.com"}, nil)
				repo.EXPECT().WasNotifiedSince(mock.Anything, kind, "Ciencias", "coord@example.com", resendSince).Return(true, nil)
			},
		},
		{
			name: "nothing pending sends nothing",
			prepare: func(repo *mocks.MockRepository, mail *mocks.MockMailClient) {
				repo.EXPECT().CountStaleCourseRequestsByFaculty(mock.Anything, createdBefore).Return(nil, nil)
				repo.EXPECT().CountStaleGroupRequestsByFaculty(mock.Anything, createdBefore).Return(nil, nil)
			},
		},
		{
			name: "faculty without coordinators is skipped",
			prepare: func(repo *mocks.MockRepository, mail *mocks.MockMailClient) {
				repo.EXPECT().CountStaleCourseRequestsByFaculty(mock.Anything, createdBefore).
					Return([]entities.FacultyPendingCount{{Faculty: entities.FacultyCiencias, Count: 1}}, nil)
				repo.EXPECT().CountStaleGroupRequestsByFaculty(mock.Anything, createdBefore).Return(nil, nil)
				repo.EXPECT().GetFacultyCoordinatorEmails(mock.Anything, entities.FacultyCiencias).Return(nil, nil)
			},
		},
		{
			name: "email failure is not logged and is reported",
			prepare: func(repo *mocks.MockRepository, mail *mocks.MockMailClient) {
				repo.EXPECT().CountStaleCourseRequestsByFaculty(mock.Anything, createdBefore).
					Return([]entities.FacultyPendingCount{{Faculty: entities.FacultyCiencias, Count: 1}}, nil)
				repo.EXPECT().CountStaleGroupRequestsByFaculty(mock.Anything, createdBefore).Return(nil, nil)
				repo.EXPECT().GetFacultyCoordinatorEmails(mock.Anything, entities.FacultyCiencias).Return([]string{"coord@example.com"}, nil)
				repo.EXPECT().WasNotifiedSince(mock.Anything, kind, "Ciencias", "coord@example.com", resendSince).Return(false, nil)
				mail.EXPECT().SendTemplate(mock.Anything, "coord@example.com", email.TemplateCoordinatorPendingReminder, mock.Anything).
					Return(errors.New("smtp down"))
			},
			wantErr: true,
		},
		{
			name: "count error",
			prepare: func(repo *mocks.MockRepository, mail *mocks.MockMailClient) {
				repo.EXPECT().CountStaleCourseRequestsByFaculty(mock.Anything, createdBefore).Return(nil, errors.New("db down"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockRepository(t)
			mail := mocks.NewMockMailClient(t)
			tt.prepare(repo, mail)

			err := newTestService(repo, mail).SendCoordinatorReminders(context.Background())
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestService_SendGroupRenewalReminders(t *testing.T) {
	dueBefore := testNow.Add(testConfig.RenewalWindow)
	kind := entities.NotificationKindGroupRenewal
	contacts := []entities.Contact{
		{Type: entities.ContactTypePhone, Value: "0212-5550000"},
		{Type: entities.ContactTypeEmail, Value: "grupo@example.com"},
	}

	tests := []struct {
		name    string
		prepare func(repo *mocks.MockRepository, mail *mocks.MockMailClient)
		wantErr bool
	}{
		{
			name: "upcoming renewal emails the group contact",
			prepare: func(repo *mocks.MockRepository, mail *mocks.MockMailClient) {
				repo.EXPECT().GetGroupsDueForRenewal(mock.Anything, dueBefore).Return([]entities.GroupRenewalDue{
					{GroupID: "7", GroupName: "Grupo Coral", RenewalDueAt: time.Date(2026, 10, 20, 9, 0, 0, 0, time.UTC)},
				}, nil)
				repo.EXPECT().GetContactsByOwner(mock.Anything, "7", entities.OwnerTypeExtensionGroup).Return(contacts, nil)
				repo.EXPECT().WasNotifiedSince(mock.Anything, kind, "7", "grupo@example.com", resendSince).Return(false, nil)
				mail.EXPECT().SendTemplate(mock.Anything, "grupo@example.com", email.TemplateGroupRenewalReminder,
					email.GroupRenewalReminderData{GroupName: "Grupo Coral", DueDate: "20/10/2026", Overdue: false}).Return(nil)
				repo.EXPECT().LogNotification(mock.Anything, entities.NotificationLog{Kind: kind, Target: "7", Recipient: "grupo@example.com"}).Return(nil)
			},
		},
		{
			name: "overdue renewal is flagged",
			prepare: func(repo *mocks.MockRepository, mail *mocks.MockMailClient) {
				repo.EXPECT().GetGroupsDueForRenewal(mock.Anything, dueBefore).Return([]entities.GroupRenewalDue{
					{GroupID: "7", GroupName: "Grupo Coral", RenewalDueAt: time.Date(2026, 9, 1, 9, 0, 0, 0, time.UTC)},
				}, nil)
				repo.EXPECT().GetContactsByOwner(mock.Anything, "7", entities.OwnerTypeExtensionGroup).Return(contacts, nil)
				repo.EXPECT().WasNotifiedSince(mock.Anything, kind, "7", "grupo@example.com", resendSince).Return(false, nil)
				mail.EXPECT().SendTemplate(mock.Anything, "grupo@example.com", email.TemplateGroupRenewalReminder,
					email.GroupRenewalReminderData{GroupName: "Grupo Coral", DueDate: "01/09/2026", Overdue: true}).Return(nil)
				repo.EXPECT().LogNotification(mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name: "group without email contact is skipped",
			prepare: func(repo *mocks.MockRepository, mail *mocks.MockMailClient) {
				repo.EXPECT().GetGroupsDueForRenewal(mock.Anything, dueBefore).Return([]entities.GroupRenewalDue{
					{GroupID: "7", GroupName: "Grupo Coral", RenewalDueAt: testNow},
				}, nil)
				repo.EXPECT().GetContactsByOwner(mock.Anything, "7", entities.OwnerTypeExtensionGroup).
					Return([]entities.Contact{{Type: entities.ContactTypePhone, Value: "0212-5550000"}}, nil)
			},
		},
		{
			name: "group reminded within the resend interval is skipped",
			prepare: func(repo *mocks.MockRepository, mail *mocks.MockMailClient) {
				repo.EXPECT().GetGroupsDueForRenewal(mock.Anything, dueBefore).Return([]entities.GroupRenewalDue{
					{GroupID: "7", GroupName: "Grupo Coral", RenewalDueAt: testNow},
				}, nil)
				repo.EXPECT().GetContactsByOwner(mock.Anything, "7", entities.OwnerTypeExtensionGroup).Return(contacts, nil)
				repo.EXPECT().WasNotifiedSince(mock.Anything, kind, "7", "grupo@example.com", resendSince).Return(true, nil)
			},
		},
		{
			name: "one failing group does not stop the others",
			prepare: func(repo *mocks.MockRepository, mail *mocks.MockMailClient) {
				repo.EXPECT().GetGroupsDueForRenewal(mock.Anything, dueBefore).Return([]entities.GroupRenewalDue{
					{GroupID: "7", GroupName: "Grupo Coral", RenewalDueAt: testNow},
					{GroupID: "8", GroupName: "Grupo Teatro", RenewalDueAt: testNow},
				}, nil)
				repo.EXPECT().GetContactsByOwner(mock.Anything, "7", entities.OwnerTypeExtensionGroup).Return(contacts, nil)
				repo.EXPECT().WasNotifiedSince(mock.Anything, kind, "7", "grupo@example.com", resendSince).Return(false, errors.New("db down"))
				repo.EXPECT().GetContactsByOwner(mock.Anything, "8", entities.OwnerTypeExtensionGroup).
					Return([]entities.Contact{{Type: entities.ContactTypeEmail, Value: "teatro@example.com"}}, nil)
				repo.EXPECT().WasNotifiedSince(mock.Anything, kind, "8", "teatro@example.com", resendSince).Return(false, nil)
				mail.EXPECT().SendTemplate(mock.Anything, "teatro@example.com", email.TemplateGroupRenewalReminder, mock.Anything).Return(nil)
				repo.EXPECT().LogNotification(mock.Anything, entities.NotificationLog{Kind: kind, Target: "8", Recipient: "teatro@example.com"}).Return(nil)
			},
			wantErr: true,
		},
		{
			name: "query error",
			prepare: func(repo *mocks.MockRepository, mail *mocks.MockMailClient) {
				repo.EXPECT().GetGroupsDueForRenewal(mock.Anything, dueBefore).Return(nil, errors.New("db down"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockRepository(t)
			mail := mocks.NewMockMailClient(t)
			tt.prepare(repo, mail)

			err := newTestService(repo, mail).SendGroupRenewalReminders(context.Background())
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestService_DeactivateExpiredGroups(t *testing.T) {
	t.Run("deactivates the expired groups", func(t *testing.T) {
		repo := mocks.NewMockRepository(t)
		repo.EXPECT().DeactivateExpiredGroups(mock.Anything, testNow).Return([]entities.GroupRenewalDue{
			{GroupID: "7", GroupName: "Grupo Coral", RenewalDueAt: testNow.Add(-time.Hour)},
		}, nil)

		assert.NoError(t, newTestService(repo, mocks.NewMockMailClient(t)).DeactivateExpiredGroups(context.Background()))
	})

	t.Run("error is reported", func(t *testing.T) {
		repo := mocks.NewMockRepository(t)
		repo.EXPECT().DeactivateExpiredGroups(mock.Anything, testNow).Return(nil, errors.New("db down"))

		assert.Error(t, newTestService(repo, mocks.NewMockMailClient(t)).DeactivateExpiredGroups(context.Background()))
	})
}
