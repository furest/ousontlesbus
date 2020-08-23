package main

var request = `SELECT trip_id
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
SELECT trip_id
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
AND CURTIME() BETWEEN begin_time - INTERVAL 1 DAY and end_time - INTERVAL 1 DAY`
