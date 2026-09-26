package entity

// BasketSubtotalBucket rounds subtotal down to the quote observation grain (D15 MVP).
const basketSubtotalBucketStepCents = 500

func BasketSubtotalBucket(subtotalCents int64) int64 {
	if subtotalCents <= 0 {
		return 0
	}
	return (subtotalCents / basketSubtotalBucketStepCents) * basketSubtotalBucketStepCents
}
