create table area_s(
	   code varchar(100) primary key,
	   name text,
	   area_m_code varchar(100),
	   area_l_code varchar(100),
	   pref_code varchar(100)
);
create table area_m(
	   code varchar(100) primary key,
	   name text
);
create table area_l(
	   code varchar(100) primary key,
	   name text
);
create table pref(
	   code varchar(100) primary key,
	   name text
);
