# bluecsv
BlueCSV is a web application for grabbing bulk data from an LDAP directory.

## What do I put in the Column # fields?
The names of LDAP attributes go in these fields. At UCSC, for instance, departmentNumber would give you the person's department. Another one would be ucscPersonPubAffiliation which says whether the person is Staff, Faculty, Undergraduate or Graduate.

## What can I upload?
Please upload only comma separated value files (.csv) when using BlueCSV. These can be made from Microsoft Excel or Numbers spreadsheets. Make sure that the unique identifiers (CruzIDs) are in the first column. Also, please make sure that the header row or first row has at least as many columns as any of the other rows.

## Data is showing up for some rows but not others
BlueCSV gets its data straight from the directory. If the ID in column 1 of the row that doesn't return results is not active in the directory, no data will be appended to that row. Likewise, if no data is avaialable for the attribute for that ID in the directory, no data will be appended to the row.
