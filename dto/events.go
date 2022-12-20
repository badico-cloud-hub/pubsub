package dto

const (
	BatchCreating                      = "batch.creating"
	BatchItemError                     = "batch-item.error"
	BatchItemsCreating                 = "batch.items-creating"
	BatchItemPixQrcodecreated          = "batch-item.pix.qrcodecreated"
	BatchItemPixUpdated                = "batch-item.pix.updated"
	BatchItemPixPaid                   = "batch-item.pix.paid"
	BatchItemPixLiquidated             = "batch-item.pix.liquidated"
	BatchItemTransferAccepted          = "batch-item.transfer.accepted"
	BatchItemTransferResolved          = "batch-item.transfer.resolved"
	BatchItemTransferRejected          = "batch-item.transfer.rejected"
	BatchItemInvoiceCreated            = "batch-item.invoice.created"
	BatchItemInvoiceInstrumentsCreated = "batch-item.invoice.instruments-created"
	BatchItemInvoicePaid               = "batch-item.invoice.paid"
	BatchItemInvoiceRejected           = "batch-item.invoice.rejected"
	ZemoGatewayChargePaid              = "zemo-gateway.charge.paid"
	ZemoGatewayChargeRejected          = "zemo-gateway.charge.rejected"
)

var AllEvents = []string{
	BatchCreating,
	BatchItemError,
	BatchItemsCreating,
	BatchItemPixQrcodecreated,
	BatchItemPixUpdated,
	BatchItemPixPaid,
	BatchItemPixLiquidated,
	BatchItemTransferAccepted,
	BatchItemTransferResolved,
	BatchItemTransferRejected,
	BatchItemInvoiceCreated,
	BatchItemInvoiceInstrumentsCreated,
	BatchItemInvoicePaid,
	BatchItemInvoiceRejected,
	ZemoGatewayChargePaid,
	ZemoGatewayChargeRejected,
}

//Events is struct for events
type Events string

func (e Events) String() string {
	switch e {
	case BatchCreating,
		BatchItemError,
		BatchItemsCreating,
		BatchItemPixQrcodecreated,
		BatchItemPixUpdated,
		BatchItemPixPaid,
		BatchItemPixLiquidated,
		BatchItemTransferAccepted,
		BatchItemTransferResolved,
		BatchItemTransferRejected,
		BatchItemInvoiceCreated,
		BatchItemInvoiceInstrumentsCreated,
		BatchItemInvoicePaid,
		BatchItemInvoiceRejected,
		ZemoGatewayChargePaid,
		ZemoGatewayChargeRejected:
		return string(e)
	default:
		return "unknown"
	}
}
