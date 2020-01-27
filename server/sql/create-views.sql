CREATE VIEW service_view AS
SELECT s.*, origin.name AS origin_name
FROM service
         JOIN origin ON origin.id = service.origin_id;


CREATE VIEW route_log_view AS
SELECT l.*,
       r.id as route_id,
       r.method,
       r.severity,
       r.path,
       r.ip,
       r.body,
       r.headers
FROM log
         LEFT JOIN route_log r ON r.log_id = log.id;