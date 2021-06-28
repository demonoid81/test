package resolvers

import (
	"github.com/sphera-erp/sphera/app"
	"github.com/sphera-erp/sphera/internal/accounting"
	"github.com/sphera-erp/sphera/internal/addresses"
	"github.com/sphera-erp/sphera/internal/contacts"
	"github.com/sphera-erp/sphera/internal/flow"
	"github.com/sphera-erp/sphera/internal/jobs"
	"github.com/sphera-erp/sphera/internal/medicalBook"
	"github.com/sphera-erp/sphera/internal/nationalities"
	"github.com/sphera-erp/sphera/internal/objectStorage"
	"github.com/sphera-erp/sphera/internal/organizations"
	"github.com/sphera-erp/sphera/internal/passports"
	"github.com/sphera-erp/sphera/internal/persons"
	"github.com/sphera-erp/sphera/internal/pingpong"
	"github.com/sphera-erp/sphera/internal/roles"
	"github.com/sphera-erp/sphera/internal/users"
)

// Resolver is
type Resolver struct {
	ping          *pingpong.Resolver
	users         *users.Resolver
	persons       *persons.Resolver
	passports     *passports.Resolver
	contacts      *contacts.Resolver
	nationalities *nationalities.Resolver
	objectStorage *objectStorage.Resolver
	medicalBooks  *medicalBook.Resolver
	addresses     *addresses.Resolver
	organizations *organizations.Resolver
	jobs          *jobs.Resolver
	jobFlow       *flow.Resolver
	accounting    *accounting.Resolver
	roles         *roles.Resolver
}

// New is
func New(app *app.App) (*Resolver, error) {
	pingResolver, _ := pingpong.NewPingResolvers(app)
	usersResolver, _ := users.NewUsersResolvers(app)
	personsResolver, _ := persons.NewPersonsResolvers(app)
	contactsResolver, _ := contacts.NewContactsResolvers(app)
	passportsResolver, _ := passports.NewPassportsResolvers(app)
	medicalBooksResolver, _ := medicalBook.NewMedicalBooksResolvers(app)
	nationalitiesResolver, _ := nationalities.NewNationalitiesResolvers(app)
	objectStorageResolver, _ := objectStorage.NewObjectStorageResolvers(app)
	addressesResolver, _ := addresses.NewAddressesResolvers(app)
	organizationsResolver, _ := organizations.NewOrganizationsResolvers(app)
	jobsResolver, _ := jobs.NewJobsResolvers(app)
	jobFlowResolver, _ := flow.NewJobFlowResolvers(app)
	accountingResolver, _ := accounting.NewAccountingResolvers(app)
	rolesResolver, _ := roles.NewRolesResolvers(app)
	return &Resolver{
		ping:          pingResolver,
		users:         usersResolver,
		persons:       personsResolver,
		contacts:      contactsResolver,
		passports:     passportsResolver,
		nationalities: nationalitiesResolver,
		objectStorage: objectStorageResolver,
		medicalBooks:  medicalBooksResolver,
		addresses:     addressesResolver,
		organizations: organizationsResolver,
		jobs:          jobsResolver,
		jobFlow:       jobFlowResolver,
		accounting:    accountingResolver,
		roles:         rolesResolver,
	}, nil
}
