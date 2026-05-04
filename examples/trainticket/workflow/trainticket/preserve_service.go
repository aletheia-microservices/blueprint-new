package trainticket

import (
	"context"
	"fmt"
	"strconv"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"github.com/google/uuid"
)

type PreserveService interface {
	Preserve(ctx context.Context, oti OrderTicketsInfo) (Order, error)
}

type PreserveServiceImpl struct {
	assuranceService AssuranceService
	basicService     BasicService
	consignService   ConsignService
	contactsService  ContactsService
	foodService      FoodService
	orderService     OrderService
	seatService      SeatService
	stationService   StationService
	//securityService  SecurityService
	travelService TravelService
	userService   UserService
	emailQueue    backend.Queue
}

func NewPreserveServiceImpl(ctx context.Context,
	assuranceService AssuranceService,
	basicService BasicService,
	consignService ConsignService,
	contactsService ContactsService,
	foodService FoodService,
	orderService OrderService,
	seatService SeatService,
	stationService StationService,
	travelService TravelService,
	userService UserService,
	queue backend.Queue,
) (PreserveService, error) {
	return &PreserveServiceImpl{
		assuranceService: assuranceService,
		basicService:     basicService,
		consignService:   consignService,
		contactsService:  contactsService,
		foodService:      foodService,
		orderService:     orderService,
		seatService:      seatService,
		stationService:   stationService,
		travelService:    travelService,
		userService:      userService,
		emailQueue:       queue,
	}, nil
}

func (p *PreserveServiceImpl) Preserve(ctx context.Context, oti OrderTicketsInfo) (Order, error) {
	// 1. Detect ticket scalper

	// 2. Query contact information
	contact, err := p.contactsService.FindContactsById(ctx, oti.ContactsID)
	if err != nil {
		return Order{}, err
	}

	// 3. Check train info and number of remaining tickets
	gtdi := TripAllDetailInfo{
		From:       oti.From,
		To:         oti.To,
		TravelDate: oti.Date,
		TripID:     oti.TripID,
	}

	gtdr, err := p.travelService.GetTripAllDetailInfo(ctx, gtdi)
	if err != nil {
		return Order{}, nil
	}

	tripResponse := gtdr.tripResponse
	if oti.SeatType == 1 { // first class
		if tripResponse.ConfortClass == 0 {
			return Order{}, fmt.Errorf("seat not enough")
		}
	} else {
		if tripResponse.EconomyClass == 0 && tripResponse.ConfortClass == 0 {
			return Order{}, fmt.Errorf("seat not enough")
		}
	}

	// 4.1. Query for travel
	travelQuery := Travel{
		Trip:          gtdr.trip,
		StartPlace:    oti.From,
		EndPlace:      oti.To,
		DepartureTime: dateToString(),
	}
	travelResult, err := p.basicService.QueryForTravel(ctx, travelQuery)
	if err != nil {
		return Order{}, err
	}

	// 4.2. Dispatch seat
	seatRequest := SeatRequest{
		TravelDate:   oti.Date,
		TrainNumber:  oti.TripID,
		StartStation: oti.From,
		DestStation:  oti.To,
		SeatType:     oti.SeatType,
		TotalNum:     travelResult.TrainType.EconomyClass,
		Stations:     travelResult.Route.Stations,
	}
	ticket, err := p.seatService.DistributeSeat(ctx, seatRequest)
	if err != nil {
		return Order{}, nil
	}

	// 4.3. Send order request
	order := Order{
		ID:                     uuid.New().String(),
		TrainNumber:            oti.TripID,
		AccountID:              oti.AccountID,
		FromStation:            oti.From,
		ToStation:              oti.To,
		BoughtDate:             dateToString(),
		ContactsName:           contact.Name,
		ContactsDocumentNumber: contact.DocumentNumber,
		DocumentType:           contact.DocumentType,
		SeatClass:              oti.SeatType,
		SeatNumber:             strconv.Itoa(ticket.SeatNo),
		Price:                  travelResult.Prices["economyClass"],
		TravelDate:             oti.Date,
	}

	cor, err := p.orderService.CreateNewOrder(ctx, order)
	if err != nil {
		return Order{}, err
	}

	// 5. Check insurance options
	if oti.Assurance != 0 {
		_, err := p.assuranceService.Create(ctx, oti.Assurance, cor.ID)
		if err != nil {
			return Order{}, err
		}
	}

	// 6. Increase the food order
	if oti.FoodType != 0 {
		foodOrder := FoodOrder{
			OrderID:  cor.ID,
			FoodType: oti.FoodType,
			FoodName: oti.FoodName,
			Price:    oti.FoodPrice,
		}

		if oti.FoodType == 2 {
			foodOrder.StationName = oti.StationName
			foodOrder.StoreName = oti.StoreName
		}

		_, err := p.foodService.CreateFoodOrder(ctx, foodOrder)
		if err != nil {
			return Order{}, nil
		}
	}

	// 7. Add consign
	if oti.ConsigneeName != "" {
		consignRequest := ConsignRequest{
			OrderID:    cor.ID,
			AccountID:  cor.AccountID,
			HandleDate: oti.HandleDate,
			TargetDate: cor.TravelDate,
			From:       cor.FromStation,
			To:         cor.ToStation,
			Consignee:  oti.ConsigneeName,
			Phone:      oti.ConsigneePhone,
			Weight:     oti.ConsigneeWeight,
			IsWithin:   oti.IsWithin,
		}
		p.consignService.InsertConsign(ctx, consignRequest)
	}

	// 8. Send notification
	user, err := p.userService.FindByUserID(ctx, cor.AccountID)
	if err != nil {
		return Order{}, err
	}
	notifyInfo := NotifyInfo{
		Email:       user.Email,
		StartPlace:  cor.FromStation,
		EndPlace:    cor.ToStation,
		Username:    user.Username,
		SeatNumber:  cor.SeatNumber,
		OrderNumber: cor.ID,
		Price:       cor.Price,
		StartTime:   cor.TravelTime,
	}
	fmt.Printf("[PRESERVE] notify info: %v\n", notifyInfo)
	/* _, err = p.emailQueue.Push(ctx, notifyInfo)
	if err != nil {
		return Order{}, err
	} */

	return cor, nil
}
