create table contact_type
(
    uuid       uuid         default gen_random_uuid() not null
        constraint "primary"
            primary key,
    name       varchar,
    created    timestamp(6) default now(),
    updated    timestamp(6),
    is_deleted bool         default false
);

create table contacts
(
    uuid              uuid         default gen_random_uuid() not null
        constraint "primary"
            primary key,
    created           timestamp(6) default now(),
    uuid_person       uuid,
    presentation      varchar,
    uuid_contact_type uuid
        constraint contacts_contact_type_uuid_fk
            references contact_type,
    updated           timestamp(6),
    is_deleted bool         default false
);

create table content
(
    uuid    uuid         default gen_random_uuid() not null
        constraint "primary"
            primary key,
    created timestamp(6) default now(),
    updated timestamp(6),
    bucket  varchar                                not null,
    is_deleted bool         default false
);

create table countries
(
    uuid       uuid    not null
        constraint "primary"
            primary key,
    country    varchar not null,
    created    timestamp(6) default now(),
    updated    timestamp(6),
    is_deleted bool         default false
);

create table medical_books
(
    uuid                            uuid         default gen_random_uuid() not null
        constraint "primary"
            primary key,
    number                          varchar,
    medical_examination_date        date,
    uuid_content_front              uuid
        constraint medical_book_content_uuid_fk
            references content,
    uuid_person                     uuid,
    created                         timestamp(6) default now(),
    updated                         timestamp(6),
    is_deleted bool         default false,
    uuid_content_back               uuid
        constraint medical_book_content_uuid_fk_2
            references content,
    have_health_restrictions        bool,
    have_medical_book               bool,
    description_health_restrictions varchar
);

create table nationalities
(
    uuid    uuid         default gen_random_uuid() not null
        constraint "primary"
            primary key,
    created timestamp(6) default now(),
    updated timestamp(6),
    is_deleted bool         default false,
    name    varchar
);

create table objects
(
    uuid       uuid         default gen_random_uuid() not null
        constraint "primary"
            primary key,
    created    timestamp(6) default now(),
    updated    timestamp(6),
    name       varchar,
    is_deleted bool         default false,
    rowid      int8         default unique_rowid()    not null
);

create table organization_positions
(
    uuid       uuid         default gen_random_uuid() not null
        constraint "primary"
            primary key,
    created    timestamp(6) default now(),
    updated    timestamp(6),
    name       varchar,
    is_deleted bool         default false
);

create table regions
(
    uuid       uuid    not null
        constraint "primary"
            primary key,
    region     varchar not null,
    created    timestamp(6) default now(),
    updated    timestamp(6),
    is_deleted bool         default false
);

create table areas
(
    uuid        uuid not null
        constraint "primary"
            primary key,
    area        varchar,
    created     timestamp(6) default now(),
    updated     timestamp(6),
    uuid_region uuid
        constraint areas_regions_uuid_fk
            references regions,
    is_deleted  bool         default false
);

create table cities
(
    uuid        uuid    not null
        constraint "primary"
            primary key,
    city        varchar not null,
    created     timestamp(6) default now(),
    updated     timestamp(6),
    uuid_region uuid
        constraint cities_regions_uuid_fk
            references regions,
    uuid_area   uuid
        constraint cities_areas_uuid_fk
            references areas,
    is_deleted  bool         default false
);

create table city_districts
(
    uuid          uuid not null
        constraint "primary"
            primary key,
    city_district varchar,
    created       timestamp(6) default now(),
    updated       timestamp(6),
    uuid_region   uuid
        constraint city_districts_regions_uuid_fk
            references regions,
    uuid_area     uuid
        constraint city_districts_areas_uuid_fk
            references areas,
    uuid_city     uuid
        constraint city_districts_cities_uuid_fk
            references cities,
    is_deleted    bool         default false
);

create table rights_to_object
(
    uuid        uuid         default gen_random_uuid() not null
        constraint "primary"
            primary key,
    created     timestamp(6) default now(),
    updated     timestamp(6),
    uuid_object uuid                                   not null,
    is_deleted  bool         default false,
    s           bool         default false,
    i           bool         default false,
    u           bool         default false,
    d           bool         default false
);

create table settlements
(
    uuid               uuid not null
        constraint "primary"
            primary key,
    settlement         varchar,
    created            timestamp(6) default now(),
    updated            timestamp(6),
    is_deleted bool         default false,
    uuid_region        uuid
        constraint settlements_regions_uuid_fk
            references regions,
    uuid_area          uuid
        constraint settlements_areas_uuid_fk
            references areas,
    uuid_city          uuid
        constraint settlements_cities_uuid_fk
            references cities,
    uuid_city_district uuid
        constraint settlements_city_districts_uuid_fk
            references city_districts
);

create table streets
(
    uuid               uuid not null
        constraint "primary"
            primary key,
    street             varchar,
    uuid_country       uuid
        constraint streets_countries_uuid_fk
            references countries,
    uuid_region        uuid
        constraint streets_regions_uuid_fk
            references regions,
    uuid_area          uuid
        constraint streets_areas_uuid_fk
            references areas,
    uuid_city          uuid
        constraint streets_cities_uuid_fk
            references cities,
    uuid_city_district uuid
        constraint streets_city_districts_uuid_fk
            references city_districts,
    uuid_settlement    uuid
        constraint streets_settlements_uuid_fk
            references settlements,
    is_deleted bool         default false,
    created            timestamp(6) default now(),
    updated            timestamp(6)
);

create table addresses
(
    uuid               uuid         default gen_random_uuid() not null
        constraint "primary"
            primary key,
    formatted_address  varchar,
    lat                float8,
    lon                float8,
    uuid_person        uuid,
    created            timestamp(6) default now(),
    updated            timestamp(6),
    is_deleted bool         default false,
    uuid_country       uuid
        constraint addresses_countries_uuid_fk
            references countries,
    uuid_region        uuid
        constraint addresses_regions_uuid_fk
            references regions,
    uuid_area          uuid
        constraint addresses_areas_uuid_fk
            references areas,
    uuid_city          uuid
        constraint addresses_cities_uuid_fk
            references cities,
    uuid_city_district uuid
        constraint addresses_city_districts_uuid_fk
            references city_districts,
    uuid_settlement    uuid
        constraint addresses_settlements_uuid_fk
            references settlements,
    uuid_street        uuid
        constraint addresses_streets_uuid_fk
            references streets,
    house              varchar,
    flat               varchar,
    uuid_organization  uuid,
    block              varchar
);

create table organizations
(
    uuid                     uuid         default gen_random_uuid() not null
        constraint "primary"
            primary key,
    created                  timestamp(6) default now(),
    updated                  timestamp(6),
    name                     varchar,
    inn                      varchar,
    kpp                      varchar,
    uuid_address_legal       uuid
        constraint organizations_addresses_uuid_fk
            references addresses,
    uuid_parent_organization uuid         default NULL
        constraint organizations_organizations_uuid_fk
            references organizations,
    uuid_departments         uuid[],
    is_deleted               bool         default false,
    uuid_address_fact        uuid
        constraint organizations_addresses_uuid_fk_2
            references addresses,
    uuid_logo                uuid
        constraint organizations_content_uuid_fk
            references content,
    prefix                   varchar,
    full_name                varchar,
    short_name               varchar,
    fee                      float8,
    uuid_persons             uuid[]
);

create table balances
(
    uuid              uuid           default gen_random_uuid() not null
        constraint "primary"
            primary key,
    created           timestamptz(6) default now(),
    updated           timestamptz(6),
    is_deleted        bool           default false,
    uuid_organization uuid
        constraint ballance_organizations_uuid_fk
            references organizations,
    amount            float8                                   not null
);

create table courses
(
    uuid              uuid         default gen_random_uuid() not null
        constraint "primary"
            primary key,
    created           timestamp(6) default now(),
    updated           timestamp(6),
    is_deleted bool         default false,
    course_type       varchar                                not null,
    content           varchar,
    uuid_organization uuid
        constraint courses_organizations_uuid_fk
            references organizations,
    name              varchar
);

create table job_types
(
    uuid                   uuid default gen_random_uuid() not null
        constraint "primary"
            primary key,
    created                timestamp(6),
    updated                timestamp(6),
    uuid_organization      uuid
        constraint job_types_organizations_uuid_fk
            references organizations,
    name                   varchar,
    uuid_courses           uuid[],
    is_deleted bool         default false,
    job_type_icon          text,
    uuid_locality_job_cost uuid[]
);

create table job_templates
(
    uuid              uuid         default gen_random_uuid() not null
        constraint "primary"
            primary key,
    created           timestamp(6) default now(),
    updated           timestamp(6),
    uuid_organization uuid
        constraint job_templates_organizations_uuid_fk
            references organizations,
    uuid_job_type     uuid
        constraint job_templates_job_types_uuid_fk
            references job_types,
    cost              float8,
    date              date,
    start_time        time(6),
    end_time          time(6),
    description       varchar,
    published         timestamp(6),
    is_deleted bool         default false,
    uuid_object       uuid
        constraint job_templates_organizations_uuid_fk_2
            references organizations,
    uuid_region       uuid
        constraint job_templates_regions_uuid_fk
            references regions,
    uuid_area         uuid
        constraint job_templates_areas_uuid_fk
            references areas,
    uuid_city         uuid
        constraint job_templates_cities_uuid_fk
            references cities,
    name              varchar,
    duration          int8
);

create table locality_job_costs
(
    uuid               uuid         default gen_random_uuid() not null
        constraint "primary"
            primary key,
    created            timestamp(6) default now(),
    updated            timestamp(6),
    uuid_organization  uuid
        constraint locality_job_costs_organizations_uuid_fk
            references organizations,
    uuid_country       uuid
        constraint locality_job_costs_countries_uuid_fk
            references countries,
    uuid_region        uuid
        constraint locality_job_costs_regions_uuid_fk
            references regions,
    uuid_area          uuid
        constraint locality_job_costs_areas_uuid_fk
            references areas,
    uuid_city          uuid
        constraint locality_job_costs_cities_uuid_fk
            references cities,
    uuid_city_district uuid
        constraint locality_job_costs_city_districts_uuid_fk
            references city_districts,
    uuid_settlement    uuid
        constraint locality_job_costs_settlements_uuid_fk
            references settlements,
    max_cost           float8,
    is_deleted         bool         default false
);

create table passports
(
    uuid                      uuid         default gen_random_uuid() not null
        constraint "primary"
            primary key,
    serial                    varchar,
    number                    varchar,
    department                varchar,
    date_issue                date,
    department_code           varchar,
    uuid_person               uuid,
    created                   timestamp(6) default now(),
    updated                   timestamp(6),
    is_deleted bool         default false,
    uuid_scan                 uuid
        constraint passports_content_uuid_fk
            references content,
    uuid_photo_registration   uuid
        constraint passports_content_uuid_fk_2
            references content,
    uuid_address_registration uuid
        constraint passports_addresses_uuid_fk
            references addresses
);

create table persons
(
    uuid                uuid         default gen_random_uuid() not null
        constraint "primary"
            primary key,
    created             timestamp(6) default now(),
    updated             timestamp(6),
    is_deleted bool         default false,
    uuid_user           uuid,
    uuid_actual_contact uuid
        constraint persons_phones_uuid_fk
            references contacts,
    uuid_passport       uuid
        constraint persons_passports_uuid_fk
            references passports,
    surname             varchar,
    name                varchar,
    patronymic          varchar,
    birth_date          date,
    gender              varchar,
    inn                 varchar,
    uuid_medical_book   uuid
        constraint persons_medical_books_uuid_fk
            references medical_books,
    uuid_country        uuid
        constraint persons_countries_uuid_fk
            references countries,
    uuid_position       uuid
        constraint persons_organization_positions_uuid_fk
            references organization_positions,
    is_contact          bool         default false,
    uuid_photo          uuid
        constraint persons_content_uuid_fk
            references content,
    recognized_fields   jsonb,
    recognize_result    jsonb,
    distance_result     jsonb,
    uuid_contacts       uuid[]
);

comment on table persons is 'Люди';

alter table addresses
    add constraint addresses_persons_uuid_fk
        foreign key (uuid_person) references persons
            on update cascade on delete cascade;

alter table contacts
    add constraint contacts_persons_uuid_fk
        foreign key (uuid_person) references persons;

create table jobs
(
    uuid              uuid         default gen_random_uuid() not null
        constraint "primary"
            primary key,
    created           timestamp(6) default now(),
    updated           timestamp(6),
    date              date,
    start_time        time(6),
    end_time          time(6),
    uuid_job_template uuid
        constraint jobs_job_templates_uuid_fk
            references job_templates,
    description       varchar,
    is_hot            bool,
    uuid_candidates   uuid[],
    uuid_executor     uuid
        constraint jobs_persons_uuid_fk
            references persons,
    uuid_statuses     uuid[],
    is_deleted bool         default false,
    uuid_object       uuid
        constraint jobs_organizations_uuid_fk
            references organizations,
    uuid_job_type     uuid
        constraint jobs_job_types_uuid_fk_2
            references job_types,
    cost              float8,
    duration          int8,
    name              varchar,
    status            varchar      default 'created',
    published         timestamp(6)
);

create table candidates
(
    uuid          uuid         default gen_random_uuid() not null
        constraint "primary"
            primary key,
    created       timestamp(6) default now(),
    updated       timestamp(6),
    uuid_person   uuid
        constraint candidates_persons_uuid_fk
            references persons,
    uuid_job      uuid
        constraint candidates_jobs_uuid_fk
            references jobs,
    candidate_tag varchar,
    is_deleted    bool         default false
);

alter table medical_books
    add constraint medical_book_persons_uuid_fk
        foreign key (uuid_person) references persons;

create table organization_contacts
(
    uuid          uuid         default gen_random_uuid() not null
        constraint "primary"
            primary key,
    created       timestamp(6) default now(),
    updated       timestamp(6),
    is_deleted bool         default false,
    uuid_person   uuid
        constraint organization_contacts_persons_uuid_fk
            references persons,
    uuid_position uuid
        constraint organization_contacts_organization_positions_uuid_fk
            references organization_positions,
    rowid         int8         default unique_rowid()    not null
);

alter table passports
    add constraint passports_persons_uuid_fk
        foreign key (uuid_person) references persons
            on update cascade on delete cascade;

create table statuses
(
    uuid         uuid         default gen_random_uuid() not null
        constraint "primary"
            primary key,
    created      timestamp(6) default now(),
    updated      timestamp(6),
    uuid_person  uuid
        constraint statuses_persons_uuid_fk
            references persons,
    uuid_job     uuid
        constraint statuses_jobs_uuid_fk
            references jobs,
    description  varchar,
    uuid_content uuid[],
    uuid_tags    uuid[],
    is_deleted   bool         default false,
    status       varchar,
    lat          float8,
    lon          float8
);

create table template_rights
(
    uuid             uuid         default gen_random_uuid() not null
        constraint "primary"
            primary key,
    created          timestamp(6) default now(),
    updated          timestamp(6),
    name             varchar,
    is_deleted       bool         default false,
    rights_to_object jsonb
);

create table users
(
    uuid               uuid         default gen_random_uuid() not null
        constraint "primary"
            primary key,
    created            timestamp(6) default now(),
    updated            timestamp(6),
    is_deleted         bool         default false,
    is_blocked         bool         default false,
    is_disabled        bool         default false,
    uuid_contact       uuid
        constraint users_contacts_uuid_fk
            references contacts,
    uuid_person        uuid
        constraint users_persons_uuid_fk
            references persons,
    type               varchar      default 'SelfEmployed'    not null,
    notification_token varchar,
    uuid_organization  uuid
        constraint users_organizations_uuid_fk
            references organizations
);

comment on table users is 'Пользователи';

alter table persons
    add constraint persons_users_uuid_fk
        foreign key (uuid_user) references users
            on update cascade on delete cascade;


---- create above / drop below ----
DROP TABLE briefings;
