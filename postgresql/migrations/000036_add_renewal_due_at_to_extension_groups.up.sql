-- Fecha en la que el grupo debe renovar su registro anual. Se fija al aprobar el grupo
-- (NOW() + 1 año) y el servicio notifier envía recordatorios cuando falta un mes o menos.
ALTER TABLE deu.extension_groups ADD COLUMN renewal_due_at TIMESTAMP DEFAULT NULL;

-- Grupos activos existentes: un año desde su última aprobación, o desde su creación si no tienen
-- una solicitud aprobada registrada.
UPDATE deu.extension_groups g
SET renewal_due_at = COALESCE(
    (
      SELECT MAX(gr.reviewed_at)
      FROM deu.group_auth_requests gr
      WHERE gr.group_id = g.id AND gr.status = 'approved'
    ),
    g.created_at
  ) + INTERVAL '1 year'
WHERE g.is_active = TRUE AND g.deleted_at IS NULL;
