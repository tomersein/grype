package v6

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/OneOfOne/xxhash"
	"gorm.io/gorm"

	"github.com/anchore/grype/internal/log"
)

func Models() []any {
	return []any{
		// core data store
		&Blob{},
		&BlobDigest{}, // only needed in write case

		// non-domain info
		&DBMetadata{},

		// data source info
		&Provider{},

		// vulnerability related search tables
		&VulnerabilityHandle{},
		&VulnerabilityAlias{},

		// package related search tables
		&AffectedPackageHandle{}, // join on package, operating system
		&OperatingSystem{},
		&OperatingSystemSpecifierOverride{},
		&Package{},
		&PackageSpecifierOverride{},

		// CPE related search tables
		&AffectedCPEHandle{}, // join on CPE
		&Cpe{},
	}
}

type ID int64

// core data store //////////////////////////////////////////////////////

type Blob struct {
	ID    ID     `gorm:"column:id;primaryKey"`
	Value string `gorm:"column:value;not null"`
}

func (b Blob) computeDigest() string {
	h := xxhash.New64()
	if _, err := h.Write([]byte(b.Value)); err != nil {
		log.Errorf("unable to hash blob: %v", err)
		panic(err)
	}
	return fmt.Sprintf("xxh64:%x", h.Sum(nil))
}

type BlobDigest struct {
	ID     string `gorm:"column:id;primaryKey"` // this is the digest
	BlobID ID     `gorm:"column:blob_id"`
	Blob   Blob   `gorm:"foreignKey:BlobID"`
}

// non-domain info //////////////////////////////////////////////////////

type DBMetadata struct {
	BuildTimestamp *time.Time `gorm:"column:build_timestamp;not null"`
	Model          int        `gorm:"column:model;not null"`
	Revision       int        `gorm:"column:revision;not null"`
	Addition       int        `gorm:"column:addition;not null"`
}

// data source info //////////////////////////////////////////////////////

// Provider is the upstream data processor (usually Vunnel) that is responsible for vulnerability records. Each provider
// should be scoped to a specific vulnerability dataset, for instance, the "ubuntu" provider for all records from
// Canonicals' Ubuntu Security Notices (for all Ubuntu distro versions).
type Provider struct {
	// Name of the Vunnel provider (or sub processor responsible for data records from a single specific source, e.g. "ubuntu")
	ID string `gorm:"column:id;primaryKey"`

	// Version of the Vunnel provider (or sub processor equivalent)
	Version string `gorm:"column:version"`

	// Processor is the name of the application that processed the data (e.g. "vunnel")
	Processor string `gorm:"column:processor"`

	// DateCaptured is the timestamp which the upstream data was pulled and processed
	DateCaptured *time.Time `gorm:"column:date_captured"`

	// InputDigest is a self describing hash (e.g. sha256:123... not 123...) of all data used by the provider to generate the vulnerability records
	InputDigest string `gorm:"column:input_digest"`
}

func (p *Provider) BeforeCreate(tx *gorm.DB) (err error) {
	// if the name and version already exist in the table then we should not insert a new record
	var existing Provider
	result := tx.Where("id = ?", p.ID).First(&existing)
	if result.Error == nil {
		if existing.Processor == p.Processor && existing.DateCaptured == p.DateCaptured && existing.InputDigest == p.InputDigest && p.Version == existing.Version {
			// record already exists
			p.ID = existing.ID
			return nil
		}

		// overwrite the existing provider if found
		existing.Processor = p.Processor
		existing.DateCaptured = p.DateCaptured
		existing.InputDigest = p.InputDigest
		existing.Version = p.Version
		if err := tx.Save(&existing).Error; err != nil {
			return fmt.Errorf("failed to update existing %q provider record: %w", p.ID, err)
		}
		return nil
	}

	// create a new provider record if not found
	return nil
}

// vulnerability related search tables //////////////////////////////////////////////////////

// VulnerabilityHandle represents the pointer to the core advisory record for a single known vulnerability from a specific provider.
type VulnerabilityHandle struct {
	ID ID `gorm:"column:id;primaryKey"`

	// Name is the unique name for the vulnerability (same as the decoded VulnerabilityBlob.ID)
	Name string `gorm:"column:name;not null;index,collate:NOCASE"`

	// Status conveys the actionability of the current record (one of "active", "analyzing", "rejected", "disputed")
	Status VulnerabilityStatus `gorm:"column:status;not null;index,collate:NOCASE"`

	// PublishedDate is the date the vulnerability record was first published
	PublishedDate *time.Time `gorm:"column:published_date;index"`

	// ModifiedDate is the date the vulnerability record was last modified
	ModifiedDate *time.Time `gorm:"column:modified_date;index"`

	// WithdrawnDate is the date the vulnerability record was withdrawn
	WithdrawnDate *time.Time `gorm:"column:withdrawn_date;index"`

	ProviderID string    `gorm:"column:provider_id;not null;index"`
	Provider   *Provider `gorm:"foreignKey:ProviderID"`

	BlobID    ID                 `gorm:"column:blob_id;index,unique"`
	BlobValue *VulnerabilityBlob `gorm:"-"`
}

func (v VulnerabilityHandle) getBlobValue() any {
	return v.BlobValue
}

func (v *VulnerabilityHandle) setBlobID(id ID) {
	v.BlobID = id
}

func (v VulnerabilityHandle) getBlobID() ID {
	return v.BlobID
}

func (v *VulnerabilityHandle) setBlob(rawBlobValue []byte) error {
	var blobValue VulnerabilityBlob
	if err := json.Unmarshal(rawBlobValue, &blobValue); err != nil {
		return fmt.Errorf("unable to unmarshal vulnerability blob value: %w", err)
	}

	v.BlobValue = &blobValue
	return nil
}

type VulnerabilityAlias struct {
	// Name is the unique name for the vulnerability
	Name string `gorm:"column:name;primaryKey;index,collate:NOCASE"`

	// Alias is an alternative name for the vulnerability that must be upstream from the Name (e.g if name is "RHSA-1234" then the upstream could be "CVE-1234-5678", but not the other way around)
	Alias string `gorm:"column:alias;primaryKey;index,collate:NOCASE;not null"`
}

// package related search tables //////////////////////////////////////////////////////

// AffectedPackageHandle represents a single package affected by the specified vulnerability. A package here is a
// name within a known ecosystem, such as "python" or "golang". It is important to note that this table relates
// vulnerabilities to resolved packages. There are cases when we have package identifiers but are not resolved to
// packages; for example, when we have a CPE but not a clear understanding of the package ecosystem and authoritative
// name (which might or might not be the product name in the CPE), in which case AffectedCPEHandle should be used.
type AffectedPackageHandle struct {
	ID              ID                   `gorm:"column:id;primaryKey"`
	VulnerabilityID ID                   `gorm:"column:vulnerability_id;index;not null"`
	Vulnerability   *VulnerabilityHandle `gorm:"foreignKey:VulnerabilityID"`

	OperatingSystemID *ID              `gorm:"column:operating_system_id;index"`
	OperatingSystem   *OperatingSystem `gorm:"foreignKey:OperatingSystemID"`

	PackageID ID       `gorm:"column:package_id;index"`
	Package   *Package `gorm:"foreignKey:PackageID"`

	BlobID    ID                   `gorm:"column:blob_id"`
	BlobValue *AffectedPackageBlob `gorm:"-"`
}

func (v AffectedPackageHandle) getBlobValue() any {
	return v.BlobValue
}

func (v *AffectedPackageHandle) setBlobID(id ID) {
	v.BlobID = id
}

func (v AffectedPackageHandle) getBlobID() ID {
	return v.BlobID
}

func (v *AffectedPackageHandle) setBlob(rawBlobValue []byte) error {
	var blobValue AffectedPackageBlob
	if err := json.Unmarshal(rawBlobValue, &blobValue); err != nil {
		return fmt.Errorf("unable to unmarshal affected package blob value: %w", err)
	}

	v.BlobValue = &blobValue
	return nil
}

// Package represents a package name within a known ecosystem, such as "python" or "golang".
type Package struct {
	ID ID `gorm:"column:id;primaryKey"`

	// Ecosystem is the tooling and language ecosystem that the package is released within
	Ecosystem string `gorm:"column:ecosystem;index:idx_package,unique,collate:NOCASE"`

	// Name is the name of the package within the ecosystem
	Name string `gorm:"column:name;index:idx_package,unique;index:idx_package_name,collate:NOCASE"`

	// CPEs is the list of Common Platform Enumeration (CPE) identifiers that represent this package
	CPEs []Cpe `gorm:"foreignKey:PackageID;constraint:OnDelete:CASCADE;"`
}

func (p *Package) BeforeCreate(tx *gorm.DB) (err error) {
	var existingPackage Package
	result := tx.Where("ecosystem = ? collate nocase AND name = ? collate nocase", p.Ecosystem, p.Name).First(&existingPackage)
	if result.Error == nil {
		// package exists; merge CPEs
		for i, newCPE := range p.CPEs {
			// if the CPE already exists, then we should use the existing record
			var existingCPE Cpe
			cpeResult := cpeWhereClause(tx, &newCPE).First(&existingCPE)
			if cpeResult.Error == nil {
				// if the record already exists, then we should use the existing record
				newCPE = existingCPE
				p.CPEs[i] = newCPE

				if existingCPE.PackageID == nil {
					log.WithFields("cpe", existingCPE, "pkg", existingPackage).Warn("CPE exists but was not associated with an already existing package until now")
					continue
				}

				if *existingCPE.PackageID != existingPackage.ID {
					return fmt.Errorf("CPE already exists for a different package (pkg=%q, existing_pkg=%q): %q", p, existingPackage, newCPE)
				}
				continue
			}

			// if the CPE does not exist, proceed with creating it
			newCPE.PackageID = &existingPackage.ID
			p.CPEs[i] = newCPE

			if err := tx.Create(&newCPE).Error; err != nil {
				return fmt.Errorf("failed to create CPE %q for package %q: %w", newCPE, existingPackage, err)
			}
		}
		// use the existing package instead of creating a new one
		*p = existingPackage
		return nil
	}

	// if the package does not exist, proceed with creating it
	for i := range p.CPEs {
		p.CPEs[i].PackageID = &p.ID
	}
	return nil
}

// PackageSpecifierOverride is a table that allows for overriding fields on v6.PackageSpecifier instances when searching for specific Packages.
type PackageSpecifierOverride struct {
	Ecosystem string `gorm:"column:ecosystem;primaryKey;index:pkg_ecosystem_idx,collate:NOCASE"`

	// below are the fields that should be used as replacement for fields in the Packages table

	ReplacementEcosystem *string `gorm:"column:replacement_ecosystem;primaryKey"`
}

// OperatingSystem represents specific release of an operating system. The resolution of the version is
// relative to the available data by the vulnerability data provider, so though there may be major.minor.patch OS
// releases, there may only be data available for major.minor.
type OperatingSystem struct {
	ID ID `gorm:"column:id;primaryKey"`

	// Name is the operating system family name (e.g. "debian")
	Name      string `gorm:"column:name;index:os_idx,unique;index,collate:NOCASE"`
	ReleaseID string `gorm:"column:release_id;index:os_idx,unique;index,collate:NOCASE"`

	// MajorVersion is the major version of a specific release (e.g. "10" for debian 10)
	MajorVersion string `gorm:"column:major_version;index:os_idx,unique;index"`

	// MinorVersion is the minor version of a specific release (e.g. "1" for debian 10.1)
	MinorVersion string `gorm:"column:minor_version;index:os_idx,unique;index"`

	// LabelVersion is an optional non-codename string representation of the version (e.g. "unstable" or for debian:sid)
	LabelVersion string `gorm:"column:label_version;index:os_idx,unique;index,collate:NOCASE"`

	// Codename is the codename of a specific release (e.g. "buster" for debian 10)
	Codename string `gorm:"column:codename;index,collate:NOCASE"`
}

func (os *OperatingSystem) VersionNumber() string {
	if os == nil {
		return ""
	}
	if os.MinorVersion != "" {
		return fmt.Sprintf("%s.%s", os.MajorVersion, os.MinorVersion)
	}
	return os.MajorVersion
}

func (os *OperatingSystem) Version() string {
	if os == nil {
		return ""
	}

	if os.LabelVersion != "" {
		return os.LabelVersion
	}

	if os.MajorVersion != "" {
		if os.MinorVersion != "" {
			return fmt.Sprintf("%s.%s", os.MajorVersion, os.MinorVersion)
		}
		return os.MajorVersion
	}

	return os.Codename
}

func (os *OperatingSystem) BeforeCreate(tx *gorm.DB) (err error) {
	// if the name, major version, and minor version already exist in the table then we should not insert a new record
	var existing OperatingSystem
	result := tx.Where("name = ? collate nocase AND major_version = ? AND minor_version = ?", os.Name, os.MajorVersion, os.MinorVersion).First(&existing)
	if result.Error == nil {
		// if the record already exists, then we should use the existing record
		*os = existing
	}
	return nil
}

// OperatingSystemSpecifierOverride is a table that allows for overriding fields on v6.OSSpecifier instances when searching for specific OperatingSystems.
type OperatingSystemSpecifierOverride struct {
	// Alias is an alternative name/ID for the operating system.
	Alias string `gorm:"column:alias;primaryKey;index:os_alias_idx,collate:NOCASE"`

	// Version is the matching version as found in the VERSION_ID field if the /etc/os-release file
	Version string `gorm:"column:version;primaryKey"`

	// VersionPattern is a regex pattern to match against the VERSION_ID field if the /etc/os-release file
	VersionPattern string `gorm:"column:version_pattern;primaryKey"`

	// Codename is the matching codename as found in the VERSION_CODENAME field if the /etc/os-release file
	Codename string `gorm:"column:codename"`

	// below are the fields that should be used as replacement for fields in the OperatingSystem table

	ReplacementName         *string `gorm:"column:replacement;primaryKey"`
	ReplacementMajorVersion *string `gorm:"column:replacement_major_version;primaryKey"`
	ReplacementMinorVersion *string `gorm:"column:replacement_minor_version;primaryKey"`
	ReplacementLabelVersion *string `gorm:"column:replacement_label_version;primaryKey"`
	Rolling                 bool    `gorm:"column:rolling;primaryKey"`
}

func (os *OperatingSystemSpecifierOverride) BeforeCreate(_ *gorm.DB) (err error) {
	if os.Version != "" && os.VersionPattern != "" {
		return fmt.Errorf("cannot have both version and version_pattern set")
	}

	return nil
}

// CPE related search tables //////////////////////////////////////////////////////

// AffectedCPEHandle represents a single CPE affected by the specified vulnerability. Note the CPEs in this table
// must NOT be resolvable to Packages (use AffectedPackageHandle for that). This table is used when the CPE is known,
// but we do not have a clear understanding of the package ecosystem or authoritative name, so we can still
// find vulnerabilities by these identifiers but not assert they are related to an entry in the Packages table.
type AffectedCPEHandle struct {
	ID              ID                   `gorm:"column:id;primaryKey"`
	VulnerabilityID ID                   `gorm:"column:vulnerability_id;not null"`
	Vulnerability   *VulnerabilityHandle `gorm:"foreignKey:VulnerabilityID"`

	CpeID ID   `gorm:"column:cpe_id"`
	CPE   *Cpe `gorm:"foreignKey:CpeID"`

	BlobID    ID                   `gorm:"column:blob_id"`
	BlobValue *AffectedPackageBlob `gorm:"-"`
}

func (v AffectedCPEHandle) getBlobID() ID {
	return v.BlobID
}

func (v AffectedCPEHandle) getBlobValue() any {
	return v.BlobValue
}

func (v *AffectedCPEHandle) setBlobID(id ID) {
	v.BlobID = id
}

func (v *AffectedCPEHandle) setBlob(rawBlobValue []byte) error {
	var blobValue AffectedPackageBlob
	if err := json.Unmarshal(rawBlobValue, &blobValue); err != nil {
		return fmt.Errorf("unable to unmarshal affected cpe blob value: %w", err)
	}

	v.BlobValue = &blobValue
	return nil
}

type Cpe struct {
	// TODO: what about different CPE versions?
	ID        ID  `gorm:"primaryKey"`
	PackageID *ID `gorm:"column:package_id;index"`

	Part            string `gorm:"column:part;not null;index:idx_cpe,unique,collate:NOCASE"`
	Vendor          string `gorm:"column:vendor;index:idx_cpe,unique,collate:NOCASE;index:idx_cpe_vendor,collate:NOCASE"`
	Product         string `gorm:"column:product;not null;index:idx_cpe,unique,collate:NOCASE;index:idx_cpe_product,collate:NOCASE"`
	Edition         string `gorm:"column:edition;index:idx_cpe,unique,collate:NOCASE"`
	Language        string `gorm:"column:language;index:idx_cpe,unique,collate:NOCASE"`
	SoftwareEdition string `gorm:"column:software_edition;index:idx_cpe,unique,collate:NOCASE"`
	TargetHardware  string `gorm:"column:target_hardware;index:idx_cpe,unique,collate:NOCASE"`
	TargetSoftware  string `gorm:"column:target_software;index:idx_cpe,unique,collate:NOCASE"`
	Other           string `gorm:"column:other;index:idx_cpe,unique,collate:NOCASE"`
}

func (c Cpe) String() string {
	parts := []string{"cpe:2.3", c.Part, c.Vendor, c.Product, c.Edition, c.Language, c.SoftwareEdition, c.TargetHardware, c.TargetSoftware, c.Other}
	for i, part := range parts {
		if part == "" {
			parts[i] = "*"
		}
	}
	return strings.Join(parts, ":")
}

func (c *Cpe) BeforeCreate(tx *gorm.DB) (err error) {
	// if the name, major version, and minor version already exist in the table then we should not insert a new record
	var existing Cpe
	result := cpeWhereClause(tx, c).First(&existing)
	if result.Error == nil {
		if c.PackageID != nil && c.PackageID != existing.PackageID {
			return fmt.Errorf("CPE already exists for a different package (pkg=%d, existing_pkg=%d): %q", c.PackageID, existing.PackageID, c)
		}

		// if the record already exists, then we should use the existing record
		*c = existing
	}
	return nil
}

func cpeWhereClause(tx *gorm.DB, c *Cpe) *gorm.DB {
	if c == nil {
		return tx
	}
	return tx.Where("part = ? AND vendor = ? AND product = ? AND edition = ? AND language = ? AND software_edition = ? AND target_hardware = ? AND target_software = ? AND other = ? collate nocase", c.Part, c.Vendor, c.Product, c.Edition, c.Language, c.SoftwareEdition, c.TargetHardware, c.TargetSoftware, c.Other)
}
