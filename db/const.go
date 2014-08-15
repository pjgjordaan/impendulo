//Copyright (c) 2013, The Impendulo Authors
//All rights reserved.
//
//Redistribution and use in source and binary forms, with or without modification,
//are permitted provided that the following conditions are met:
//
//  Redistributions of source code must retain the above copyright notice, this
//  list of conditions and the following disclaimer.
//
//  Redistributions in binary form must reproduce the above copyright notice, this
//  list of conditions and the following disclaimer in the documentation and/or
//  other materials provided with the distribution.
//
//THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
//ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
//WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
//DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR
//ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
//(INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
//LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON
//ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
//(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
//SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package db

const (
	//Mongodb collection name.
	USERS       = "users"
	SUBMISSIONS = "submissions"
	ASSIGNMENTS = "assignments"
	FILES       = "files"
	RESULTS     = "results"
	TESTS       = "tests"
	PROJECTS    = "projects"
	SKELETONS   = "skeletons"
	JPF         = "jpf"
	PMD         = "pmd"
	MAKE        = "make"
	//Mongodb command
	SET    = "$set"
	OR     = "$or"
	AND    = "$and"
	NOT    = "$not"
	NOR    = "$nor"
	LT     = "$lt"
	LTE    = "$lte"
	GT     = "$gt"
	GTE    = "$gte"
	IN     = "$in"
	NE     = "$ne"
	NIN    = "$nin"
	EXISTS = "$exists"
	ISTYPE = "$type"
	//Mongodb connection and db names
	ADDRESS      = "mongodb://localhost/"
	DEFAULT_DB   = "impendulo"
	DEBUG_DB     = "impendulo_debug"
	TEST_DB      = "impendulo_test"
	BACKUP_DB    = "impendulo_backup"
	DEFAULT_CONN = ADDRESS + DEFAULT_DB
	DEBUG_CONN   = "mongodb://localhost/impendulo_debug"
	TEST_CONN    = "mongodb://localhost/impendulo_test"
	//Field names
	TARGET       = "target"
	ID           = "_id"
	PROJECTID    = "projectid"
	ASSIGNMENTID = "assignmentid"
	PROJECT      = "project"
	USER         = "user"
	TIME         = "time"
	TYPE         = "type"
	TEST         = "test"
	NAME         = "name"
	PKG          = "package"
	LANG         = "lang"
	SUBID        = "subid"
	INFO         = "info"
	DATA         = "data"
	FILEID       = "fileid"
	TESTID       = "testid"
	SKELETON     = "skeleton"
	REPORT       = "report"
	STATUS       = "status"
	PWORD        = "password"
	SALT         = "salt"
	ACCESS       = "access"
	DESCRIPTION  = "description"
	COMMENTS     = "comments"
	TESTCASES    = "testcases"
	START        = "start"
	END          = "end"
)
