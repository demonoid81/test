#!/bin/sh
/app/wfmt server --tarantoolUrl=http://tarantool --db.host=roach


query Jobs($filter: JobFilter, $sort: [JobSort!]) {
   jobs: activeJobs(filter: $filter, sort: $sort) {
     uuid
     cost
     description
     startTime
     endTime
     duration
     name
     date
     jobType {
       name
       icon
     }
     candidates {
       uuid
     }
     executor {
       uuid
     }
     object {
       name
       fullName
       parentOrganization {
         name
         fullName
         shortName
         logo {
           bucket
           uuid
         }
       }
       logo {
         bucket
         uuid
       }
       addressFact {
         lat
         lon
         formattedAddress
       }
     }
     isHot
   }
 }variables:  {filter:{date:{gte:2021-06-09] isHot:false object:{addressFact:{and:[{lon:{between:[37.42287626988585 37.43397102684307]]] {lat:{between:[55.6271222345989 55.63671136551363]]]]]]] sort:[{field:date order:ASC]]]
