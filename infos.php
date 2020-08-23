<?php
$servername = "localhost";
$username = "tec";
$password = "tecpassword";
$dbname = "TEC";

$conn = new mysqli($servername, $username, $password, $dbname);

if($conn->connect_error){
	die("Connection do DB failed :" . $conn->connect_error);
}
$query = '
SELECT current_trips.trip_id as trip_id, 
       current_trips.begin_time as begin_time, 
       current_trips.end_time as end_time, 
       trips.route_id as route_id,
       trips.service_id as service_id
FROM
(
	SELECT trip_id, begin_time, end_time
	FROM trip_times 
	WHERE trip_id IN (  SELECT trip_id 
                        FROM trips 
                        WHERE service_id IN (   SELECT service_id 
                                                FROM calendar 
                                                WHERE CURDATE() BETWEEN start_date AND end_date 
                                                AND ( (DAYNAME(CURDATE()) = "Monday" AND monday = TRUE) 
                                                    OR (DAYNAME(CURDATE()) = "Tuesday" AND tuesday = TRUE) 
                                                    OR (DAYNAME(CURDATE()) = "Wednesday" AND wednesday = TRUE) 
                                                    OR (DAYNAME(CURDATE()) = "Thursday" AND thursday = TRUE) 
                                                    OR (DAYNAME(CURDATE()) = "Friday" AND friday = TRUE) 
                                                    OR (DAYNAME(CURDATE()) = "Saturday" AND saturday = TRUE) 
                                                    OR (DAYNAME(CURDATE()) = "Sunday" AND sunday = TRUE) 
                                                    ) 
                                                AND service_id NOT IN ( SELECT service_id 
                                                                        FROM calendar_dates 
                                                                        WHERE CURDATE() = date 
                                                                        AND exception_type = 2 
                                                                        )  
                                                UNION 
                                                SELECT service_id 
                                                FROM calendar_dates 
                                                WHERE CURDATE() = date 
                                                AND exception_type = 1 
                                            ) 
                    ) 
	AND CURTIME() BETWEEN begin_time and end_time
	UNION ALL
	SELECT trip_id, begin_time, end_time
	FROM trip_times 
	WHERE trip_id IN (  SELECT trip_id 
                        FROM trips 
                        WHERE service_id IN (   SELECT service_id 
                                                FROM calendar 
                                                WHERE CURDATE() - INTERVAL 1 DAY BETWEEN start_date AND end_date 
                                                AND ( (DAYNAME(CURDATE() - INTERVAL 1 DAY) = "Monday" AND monday = TRUE) 
                                                    OR (DAYNAME(CURDATE() - INTERVAL 1 DAY) = "Tuesday" AND tuesday = TRUE) 
                                                    OR (DAYNAME(CURDATE() - INTERVAL 1 DAY) = "Wednesday" AND wednesday = TRUE) 
                                                    OR (DAYNAME(CURDATE() - INTERVAL 1 DAY) = "Thursday" AND thursday = TRUE) 
                                                    OR (DAYNAME(CURDATE() - INTERVAL 1 DAY) = "Friday" AND friday = TRUE) 
                                                    OR (DAYNAME(CURDATE() - INTERVAL 1 DAY) = "Saturday" AND saturday = TRUE) 
                                                    OR (DAYNAME(CURDATE() - INTERVAL 1 DAY) = "Sunday" AND sunday = TRUE) 
                                                    ) 
                                                AND service_id NOT IN ( SELECT service_id 
                                                                        FROM calendar_dates 
                                                                        WHERE CURDATE() - INTERVAL 1 DAY = date 
                                                                        AND exception_type = 2 
                                                                        )  
                                                UNION 
                                                SELECT service_id 
                                                FROM calendar_dates 
                                                WHERE CURDATE() - INTERVAL 1 DAY = date 
                                                AND exception_type = 1 
                                            ) 
                    ) 
	AND CURTIME() BETWEEN begin_time - INTERVAL 1 DAY and end_time - INTERVAL 1 DAY
) as current_trips
INNER JOIN trips as trips ON (trips.trip_id = current_trips.trip_id)';

$trips = $conn->query($query);

$nbTrips = $trips->num_rows;

echo("Nombre de bus actuellement en circulation : " . $nbTrips);
echo("</br>");
echo("Liste des trips : ");

//current_trips.trip_id as trip_id, 
//       current_trips.begin_time as begin_time, 
//       current_trips.end_time as end_time, 
//       trips.route_id as route_id,
//       trips.service_id as service_id
 echo("<table><tr><th>trip_id</th><th>begin_time</th><th>end_time</th><th>route_id</th><th>service_id</th></tr>");
while($row = $trips->fetch_assoc()){
	echo("<tr><td>".$row['trip_id']."</td><td>".$row['begin_time']."</td><td>".$row['end_time']."</td><td>".$row['route_id']."</td><td>".$row['service_id']."</td></tr>");
}
echo "</table>";
