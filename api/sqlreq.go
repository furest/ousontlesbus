package main

var SQL_REQUEST = `SELECT svcs.trip_id,
    svcs.begin_time,
    svcs.end_time,
    r.route_id,
    r.route_short_name,
    r.route_long_name
FROM (
        SELECT t.trip_id,
            tt.begin_time,
            tt.end_time,
            t.route_id
        FROM (
                (
                    SELECT c.service_id
                    FROM calendar c
                    WHERE CURDATE() BETWEEN c.start_date and c.end_date
                        AND (
                            (
                                DAYOFWEEK(CURDATE()) = 2
                                AND monday = TRUE
                            )
                            OR (
                                DAYOFWEEK(CURDATE()) = 3
                                AND tuesday = TRUE
                            )
                            OR (
                                DAYOFWEEK(CURDATE()) = 4
                                AND wednesday = TRUE
                            )
                            OR (
                                DAYOFWEEK(CURDATE()) = 5
                                AND thursday = TRUE
                            )
                            OR (
                                DAYOFWEEK(CURDATE()) = 6
                                AND friday = TRUE
                            )
                            OR (
                                DAYOFWEEK(CURDATE()) = 7
                                AND saturday = TRUE
                            )
                            OR (
                                DAYOFWEEK(CURDATE()) = 1
                                AND sunday = TRUE
                            )
                        )
                )
                UNION
                (
                    SELECT cd.service_id
                    FROM calendar_dates cd
                    WHERE CURDATE() = cd.date
                        AND cd.exception_type = 1
                )
                EXCEPT (
                        SELECT cd.service_id
                        FROM calendar_dates cd
                        WHERE CURDATE() = cd.date
                            AND cd.exception_type = 2
                    )
            ) svc
            INNER JOIN trips t ON svc.service_id = t.service_id
            INNER JOIN trip_times tt ON t.trip_id = tt.trip_id
        WHERE CURTIME() BETWEEN tt.begin_time AND DATE_ADD(tt.end_time, INTERVAL ? MINUTE)
        UNION
        SELECT t.trip_id,
            tt.begin_time,
            tt.end_time,
            t.route_id
        FROM (
                (
                    SELECT c.service_id
                    FROM calendar c
                    WHERE CURDATE() - INTERVAL 1 DAY BETWEEN c.start_date and c.end_date
                        AND (
                            (
                                DAYOFWEEK(CURDATE() - INTERVAL 1 DAY) = 2
                                AND monday = TRUE
                            )
                            OR (
                                DAYOFWEEK(CURDATE() - INTERVAL 1 DAY) = 3
                                AND tuesday = TRUE
                            )
                            OR (
                                DAYOFWEEK(CURDATE() - INTERVAL 1 DAY) = 4
                                AND wednesday = TRUE
                            )
                            OR (
                                DAYOFWEEK(CURDATE() - INTERVAL 1 DAY) = 5
                                AND thursday = TRUE
                            )
                            OR (
                                DAYOFWEEK(CURDATE() - INTERVAL 1 DAY) = 6
                                AND friday = TRUE
                            )
                            OR (
                                DAYOFWEEK(CURDATE() - INTERVAL 1 DAY) = 7
                                AND saturday = TRUE
                            )
                            OR (
                                DAYOFWEEK(CURDATE() - INTERVAL 1 DAY) = 1
                                AND sunday = TRUE
                            )
                        )
                )
                UNION
                (
                    SELECT cd.service_id
                    FROM calendar_dates cd
                    WHERE CURDATE() - INTERVAL 1 DAY = cd.date
                        AND cd.exception_type = 1
                )
                EXCEPT (
                        SELECT cd.service_id
                        FROM calendar_dates cd
                        WHERE CURDATE() - INTERVAL 1 DAY = cd.date
                            AND cd.exception_type = 2
                    )
            ) svc
            INNER JOIN trips t ON svc.service_id = t.service_id
            INNER JOIN trip_times tt ON t.trip_id = tt.trip_id
        WHERE CURTIME() + INTERVAL 1 DAY BETWEEN tt.begin_time AND DATE_ADD(tt.end_time, INTERVAL ? MINUTE)
    ) svcs
    INNER JOIN routes r ON svcs.route_id = r.route_id`
