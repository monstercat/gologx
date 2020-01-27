CREATE VIEW service_view AS
SELECT s.*, origin.name AS origin_name
FROM service
         JOIN origin ON origin.id = service.origin_id;