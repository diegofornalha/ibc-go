package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/ibc-go/v3/modules/apps/29-fee/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	ibctesting "github.com/cosmos/ibc-go/v3/testing"
)

func (suite *KeeperTestSuite) TestRegisterCounterpartyAddress() {
	var (
		sender       string
		counterparty string
	)

	testCases := []struct {
		name     string
		expPass  bool
		malleate func()
	}{
		{
			"success",
			true,
			func() {},
		},
		{
			"counterparty is an arbitrary string",
			true,
			func() { counterparty = "arbitrary-string" },
		},
	}

	for _, tc := range testCases {
		suite.SetupTest()
		ctx := suite.chainA.GetContext()

		sender = suite.chainA.SenderAccount.GetAddress().String()
		counterparty = suite.chainB.SenderAccount.GetAddress().String()
		tc.malleate()
		msg := types.NewMsgRegisterCounterpartyAddress(sender, counterparty, ibctesting.FirstChannelID)

		_, err := suite.chainA.SendMsgs(msg)

		if tc.expPass {
			suite.Require().NoError(err) // message committed

			counterpartyAddress, _ := suite.chainA.GetSimApp().IBCFeeKeeper.GetCounterpartyAddress(ctx, suite.chainA.SenderAccount.GetAddress().String(), ibctesting.FirstChannelID)
			suite.Require().Equal(counterparty, counterpartyAddress)
		} else {
			suite.Require().Error(err)
		}
	}
}

func (suite *KeeperTestSuite) TestPayPacketFee() {
	testCases := []struct {
		name     string
		expPass  bool
		malleate func()
	}{
		{
			"success",
			true,
			func() {},
		},
		{
			"fee module is locked",
			false,
			func() {
				lockFeeModule(suite.chainA)
			},
		},
	}

	for _, tc := range testCases {
		suite.SetupTest()
		suite.coordinator.Setup(suite.path) // setup channel

		refundAcc := suite.chainA.SenderAccount.GetAddress()
		channelID := suite.path.EndpointA.ChannelID
		fee := types.Fee{
			RecvFee:    defaultRecvFee,
			AckFee:     defaultAckFee,
			TimeoutFee: defaultTimeoutFee,
		}
		msg := types.NewMsgPayPacketFee(fee, suite.path.EndpointA.ChannelConfig.PortID, channelID, refundAcc.String(), []string{})

		tc.malleate()

		_, err := suite.chainA.GetSimApp().IBCFeeKeeper.PayPacketFee(sdk.WrapSDKContext(suite.chainA.GetContext()), msg)

		if tc.expPass {
			suite.Require().NoError(err) // message committed
		} else {
			suite.Require().Error(err)
		}
	}
}

func (suite *KeeperTestSuite) TestPayPacketFeeAsync() {
	testCases := []struct {
		name     string
		expPass  bool
		malleate func()
	}{
		{
			"success",
			true,
			func() {},
		},
		{
			"fee module is locked",
			false,
			func() {
				lockFeeModule(suite.chainA)
			},
		},
	}

	for _, tc := range testCases {
		suite.SetupTest()
		suite.coordinator.Setup(suite.path) // setup channel

		ctxA := suite.chainA.GetContext()

		refundAcc := suite.chainA.SenderAccount.GetAddress()

		// build packetID
		channelID := suite.path.EndpointA.ChannelID
		fee := types.Fee{
			RecvFee:    defaultRecvFee,
			AckFee:     defaultAckFee,
			TimeoutFee: defaultTimeoutFee,
		}
		seq, _ := suite.chainA.App.GetIBCKeeper().ChannelKeeper.GetNextSequenceSend(ctxA, suite.path.EndpointA.ChannelConfig.PortID, suite.path.EndpointA.ChannelID)

		// build fee
		packetID := channeltypes.NewPacketId(suite.path.EndpointA.ChannelConfig.PortID, channelID, seq)
		packetFee := types.NewPacketFee(fee, refundAcc.String(), nil)

		tc.malleate()

		msg := types.NewMsgPayPacketFeeAsync(packetID, packetFee)
		_, err := suite.chainA.GetSimApp().IBCFeeKeeper.PayPacketFeeAsync(sdk.WrapSDKContext(suite.chainA.GetContext()), msg)

		if tc.expPass {
			suite.Require().NoError(err) // message committed
		} else {
			suite.Require().Error(err)
		}
	}
}
